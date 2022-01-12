[![GoDoc](http://godoc.org/github.com/mayowa/xtemplate?status.svg)](http://godoc.org/github.com/mayowa/xtemplate)

# xtemplate

a wrapper around text/template inspired by:

* https://github.com/dannyvankooten/extemplate
* https://github.com/tyler-sommer/stick

## Overview

xtemplate wraps text/template to provide:

* file based inheritance
* optional caching
* syntax sugar on function/method calls (eg fn(arg1, arg2) instead of fn arg1 arg2)
* syntax sugar on template calls
* VueJs like components

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

## Syntax sugar

### C like function calls

```html

{{ login("name", "password") }}

instead of
{{ login "name" "password" }}

```

## make template calls look like calling functions

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
  <input>
</tag>

gets translated to a function call to a user defined tag template function
which then generates the output
tag (type, attributes, content)
```

## Components

The component feature transforms <component type="card">
a template with the same name as the value of the type attribute must exist in the "_components" subfolder

### Slots

A component can define a slot which can be overridden when it is called

```html

<component type="card">
  <slot name="title">A title</slot>
</component>
```

if a component defines a default slot the content of the component tag will be placed within the components default slot

```html

<component type="card">
  this will go in cards default slot if card defines one
</component>
```

### Component Templates

Component templates are valid text/template file with special semantics

```html

<div class="box">
  {{block "#slot--default" .}}
  default content
  {{end}}
</div>
```

```html

<component type="box">
  lorem ipsum
</component>
```

### Context and Attributes

The context passed into a component is of type map[string]interface{} the parent context is stored in '
ctx' (`.ctx.Name`) and component attributes such id|class|style are stored in 'props' (`.props.style`)

```html

<component type="box" class="red">
  lorem ipsum
</component>
```

```html

<div class="box {{$.props.class}}">
  {{block "#slot--default" .}}
  default content
  {{end}}
</div>
```
