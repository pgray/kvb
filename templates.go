package main

var rootHTML = `
<!DOCTYPE html>
<html lang="en">
  <body style="margin-left: auto;margin-right:auto;width:60%;">
    <h1>Sections:</h1>
    {{ range $key, $value := .Sections }}
    <a href="b/{{ $value }}">{{ $key }} : {{ $value }}</a>
    {{ end }}
  </body>
</html>
`

var browseHTML = `
<!DOCTYPE html>
<html lang="en">
  <body style="margin-left: auto;margin-right:auto;width:60%;">
    <h1><a href="/">/</a> BROWSE</h1>
    <h1>{{.Section}} / {{.Title}}</h1>
    {{.Body}}
  </body>
</html>
`

var editHTML = `
<!DOCTYPE html>
<html lang="en">
  <body style="margin-left: auto; margin-right:auto;width:60%;">
    <h1>EDIT {{.Section}} / {{.Title}}</h1>
    <form action="/s/{{.Section}}/{{.Title}}" method="POST">
    <div><textarea name="body" rows="30" style="width: 100%;">{{printf "%s" .Body}}</textarea></div>
    <div><input type="submit" value="Save"></div>
    </form>
  </body>
</html>
`

var sectionHTML = `
<!DOCTYPE html>
<html lang="en">
  <body style="margin-left: auto;margin-right:auto;width:60%;">
    <h1>{{ .Section }} Posts:</h1>
    {{ $section := .Section }}
    {{ range $value := .Posts }}
    <a href="{{ $section }}\/{{ $value }}">{{ $value }}</a>
    {{ end }}
  </body>
</html>
`
