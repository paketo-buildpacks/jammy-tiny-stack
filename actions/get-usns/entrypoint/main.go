package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"
)

var distroToVersionRegex map[string]string = map[string]string{
	"jammy":  `22\.04`,
	"focal":  `20\.04`,
	"bionic": `18\.04`,
}

type USN struct {
	AffectedPackages []string `json:"affected_packages"`
	CVEs             []CVE    `json:"cves"`
	Title            string   `json:"title"`
	URL              string   `json:"url"`
}

type CVE struct {
	Description string `json:"description"`
	Title       string `json:"title"`
	URL         string `json:"url"`
}

func main() {
	var config struct {
		Distro       string
		LastUSNsJSON string
		Output       string
		PackagesJSON string
		RSSURL       string
	}
	// TODO: Add some tests :D

	flag.StringVar(&config.LastUSNsJSON,
		"last-usns",
		`[]`,
		"JSON array of last known USNs")
	flag.StringVar(&config.RSSURL,
		"feed-url",
		"https://ubuntu.com/security/notices/rss.xml",
		"URL of RSS feed")
	flag.StringVar(&config.PackagesJSON,
		"packages",
		`[]`,
		"JSON array of relevant packages")
	flag.StringVar(&config.Distro,
		"distro",
		`bionic`,
		"Name of Ubuntu distro: jammy, bionic")
	flag.StringVar(&config.Output,
		"output",
		"",
		"Path to output JSON file")

	flag.Parse()

	var lastUSNs []USN
	if config.LastUSNsJSON != "" {
		err := json.Unmarshal([]byte(config.LastUSNsJSON), &lastUSNs)
		if err != nil {
			log.Fatal(err)
		}
	}

	var packages []string
	if config.PackagesJSON != "" {
		err := json.Unmarshal([]byte(config.PackagesJSON), &packages)
		if err != nil {
			log.Fatal(err)
		}
	}

	newUSNs, err := getNewUSNsFromFeed(config.RSSURL, lastUSNs, distroToVersionRegex[config.Distro])
	if err != nil {
		log.Fatal(err)
	}

	filtered := filterUSNsByPackages(newUSNs, packages)

	output, err := json.Marshal(filtered)
	if err != nil {
		log.Fatal(err)
	}

	if len(filtered) == 0 {
		output = []byte(`[]`)
	}

	fmt.Printf("::set-output name=usns::%s\n", string(output))

	if config.Output != "" {
		path, err := filepath.Abs(config.Output)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(path, output, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func filterUSNsByPackages(usns []USN, packages []string) (filtered []USN) {
	if len(packages) == 0 {
		fmt.Println("No packages specified. Skipping filtering.")
		return usns
	}

	for _, usn := range usns {
	matchPkgs:
		for _, affected := range usn.AffectedPackages {
			for _, pkg := range packages {
				if pkg == affected {
					filtered = append(filtered, usn)
					break matchPkgs
				}
			}
		}
	}
	return filtered
}

func getNewUSNsFromFeed(rssURL string, lastUSNs []USN, distro string) ([]USN, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(rssURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing rss feed: %w", err)
	}

	fmt.Println("Looking for new USNs:")
	var feedUSNs []USN
	for _, item := range feed.Items {
		if (len(lastUSNs) > 0) && item.Title == lastUSNs[0].Title {
			// We've already seen this USN
			// No need to keep adding to list

			fmt.Println("No more new USNs.")
			break
		}
		fmt.Printf("New USN found: %s\n", item.Title)

		usnURL, err := url.Parse(item.Link)
		if err != nil {
			return nil, fmt.Errorf("error parsing URL of USN %s: %w", item.Title, err)
		}

		usnBody, code, err := get(usnURL.String())
		if err != nil || code != http.StatusOK {
			return nil, fmt.Errorf("error getting USN: %w", err)
		}

		cves, err := extractCVEs(usnBody, usnURL)
		if err != nil {
			return nil, fmt.Errorf("error extracting CVEs from USN %s: %w", item.Title, err)
		}

		feedUSNs = append(feedUSNs, USN{
			Title:            item.Title,
			URL:              item.Link,
			CVEs:             cves,
			AffectedPackages: getAffectedPackages(usnBody, distro),
		})
	}

	return feedUSNs, nil
}

func getAffectedPackages(usnBody, versionRegex string) []string {
	re := regexp.MustCompile("Update instructions</h2>(.*?)References")
	packagesList := re.FindString(usnBody)

	re = regexp.MustCompile(fmt.Sprintf(`%s.*?</ul>`, versionRegex)) /// TODO: Get affected packages for a specific version of UBUNTU
	bionicPackages := re.FindString(packagesList)

	re = regexp.MustCompile(`<li class="p-list__item">(.*?)</li>`)
	listMatches := re.FindAllStringSubmatch(bionicPackages, -1)

	packages := make([]string, 0)
	for _, listItem := range listMatches {
		packages = append(packages, getPackageNameFromHTML(strings.TrimSpace(listItem[1])))
	}

	return packages
}

func get(url string) (string, int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", 0, err
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}

	body := html.UnescapeString(string(respBody))
	body = strings.ReplaceAll(body, "\n", " ")
	body = strings.ReplaceAll(body, "<br />", " ")
	body = strings.ReplaceAll(body, "<br>", " ")
	body = strings.ReplaceAll(body, "</br>", " ")

	return body, resp.StatusCode, nil
}
func extractCVEs(usnBody string, usnURL *url.URL) ([]CVE, error) {

	// matches <a href="/security/CVE-2022-1664">CVE-2022-1664</a> or
	// <a href="/cve/CVE-2022-1664">CVE-2022-1664</a>
	re := regexp.MustCompile(`<a.*?href="([\S]*?(:?cve|security)\/CVE.*?)">(.*?)<\/a.*?>`)

	cves := re.FindAllStringSubmatch(usnBody, -1)
	// if len(cves) > 0 {
	// 	for i, match := range cves[0] {
	// 		fmt.Println("match ", i, ": ", match)
	// 	}
	// }

	re = regexp.MustCompile(`.*?href="([\S]*?launchpad\.net/bugs.*?)">(.*?)</li`)
	lps := re.FindAllStringSubmatch(usnBody, -1)

	var cveArray []CVE
	for _, cve := range cves {
		cveURL := url.URL{
			Scheme: "https",
			Host:   usnURL.Hostname(),
			Path:   cve[1],
		}

		description, err := getCVEDescription(cveURL.String())
		if err != nil {
			return nil, fmt.Errorf("error getting description for CVE %s: %w", cve[2], err)
		}

		cveArray = append(cveArray, CVE{
			Title:       cve[3],
			URL:         cveURL.String(),
			Description: description,
		})
	}

	for _, lp := range lps {
		description, err := getLPDescription(lp[1])
		if err != nil {
			return nil, fmt.Errorf("error getting description for launchpad bug %s: %w", lp[2], err)
		}

		cveArray = append(cveArray, CVE{
			Title:       lp[2],
			URL:         lp[1],
			Description: description,
		})
	}

	return cveArray, nil
}

func getCVEDescription(url string) (string, error) {
	body, code, err := get(url)
	if err != nil {
		return "", err
	}

	if code != http.StatusOK {
		return "", nil
	}

	re := regexp.MustCompile(`Published: <strong.*?<p>(.*?)</p>`)
	desc := re.FindStringSubmatch(body)
	if len(desc) >= 2 {
		description := desc[1]
		return strings.TrimSpace(description), nil
	}

	return "", nil
}

func getLPDescription(url string) (string, error) {
	body, code, err := get(url)
	if err != nil {
		return "", err
	}

	if code != http.StatusOK {
		return "", nil
	}

	re := regexp.MustCompile(`"edit-title">.*?<span.*?>(.*?)</span>`)
	title := re.FindStringSubmatch(body)
	return strings.TrimSpace(title[1]), nil
}

func getPackageNameFromHTML(listItem string) string {
	if strings.HasPrefix(listItem, "<a href=") {
		re := regexp.MustCompile(`<a href=".*?">(.*?)</a>`)
		packageMatch := re.FindStringSubmatch(listItem)
		return packageMatch[1]
	}
	return strings.Split(listItem, " ")[0]
}
