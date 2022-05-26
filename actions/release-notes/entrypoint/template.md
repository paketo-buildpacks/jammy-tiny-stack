## Images
Build: `{{- .BuildImage -}}`
Run: `{{- .RunImage -}}`
{{- if ne (len .PatchedArray) 0 }}

## Patched USNs
{{ range .PatchedArray }}
- [{{- .Title -}}]({{- .URL -}})
{{- end }}
{{- end }}

## Build Package Diff
{{ if eq .BuildDiff "" }}
No diff in build image packages.
{{- else }}
```diff
{{ .BuildDiff }}
```
{{ end }}

## Run Package Diff
{{ if eq .RunDiff "" }}
No diff in run image packages.

{{ else }}
```diff
{{ .RunDiff }}
```
{{ end }}
