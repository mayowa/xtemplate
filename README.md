[![GoDoc](http://godoc.org/github.com/mayowa/xtemplate?status.svg)](http://godoc.org/github.com/mayowa/xtemplate)

# xtemplate
a wrapper around text/template
inspired by:
* https://github.com/dannyvankooten/extemplate
* https://github.com/tyler-sommer/stick

## Overview
xtemplate wraps text/template to provide:
* file based inheritance
* optional caching
* syntax sugar on function/method calls (eg fn(arg1, arg2) instead of fn arg1 arg2)
* syntax sugar on template calls


```html
<!-- master.html -->
<html>
<body>
    <h1>
        Hello from:
        {{block "source" }} master {{end}}
    </h1>
</body>
</html>
```


```html
<!-- extras.html -->
{{ define "footer" }} footer {{end}}
```

```html
<!-- child.html -->
{{ extends "master.html" }}
{{ include "extra.html" }}

{{ define "source" }}
  child
  {{ template footer }}
{{end}}
```

rendering child.html will output
```
Hello from child footer
```

## Syntax sugar: c like function calls
```html

{{ login("name", "password") }}

instead of
{{ login "name" "password" }}

```

## Syntax sugar: template calls
use defined template blocks (imported using the include directive)

```html
{{define "hello"}}
  Hello {{ .name }}, age: {{.age}}
{{end}}

{{define "button"}}
  a {{ upper("button") }} {{ . }}
{{end}}

call the blocks as follows

{{ macro button("submit") }}

or

{{ template button("submit") }}

or

{{ macro hello("name"::"ayo", "age"::18) }}

```

## Syntax sugar: custom tags
```html
<tag type="field" label="name">
  <input >
</tag>

gets translated to a function call to a user defined tag template function
which then generates the output
tag (type, attributes, content)
```
