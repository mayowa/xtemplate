# xtemplate
a wrapper around text/template
inspired by:
 https://github.com/dannyvankooten/extemplate
 https://github.com/tyler-sommer/stick

## Overview
xtemplate wraps text/template to provide:
* file based inheritance
* optional caching
* syntax sugar on function/method calls (eg fn(arg1, arg2) instead of fn arg1 arg2)
* syntax sugar on template calls

### extends
makes file based inheritance possible

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

{{ include "file.html" }}

and functions can be written "normally"