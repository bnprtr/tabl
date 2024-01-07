# tabl

[![Go Reference](https://pkg.go.dev/badge/github.com/bnprtr/tabl.svg)](https://pkg.go.dev/github.com/bnprtr/tabl)

[Templ](https://templ.guide) Table Component Generator

## Generating Code

Tabl will parse your go or templ files and look for any struct definition that matches
one of the regexp patterns passed in as positional argument then a table component 
will be rendered for that data type.

The code generated will be written out to a file matching the name of the file
parsed with the exception of the suffix being replaced with `_tabl.templ`. For example,
`person.go` code would be generated into `person_tabl.templ`.

### Examples

The following example would generate components for the `Person` and `Location` data
types and write the tabl code to `view_models_tabl.templ`.

```sh
tabl -file view_models.go Person Location
```

The following example would generate components for all data types in the file
and write the tabl code to `view_models_tabl.templ`.

```sh
tabl -file view_models.go .*
```

It's also fine to use go generate!

```go
//go:generate go run github.com/bnprtr/tabl@latest -file views.go
package views

type DataType struct {
  Property1 string
  Property2 int
}
```

## Data Types

Tabl will parse the targeted struct definition and extract some information from
it for code generation, such as the field names. Additionally, some additional
parameters may be defined in the struct tags. the `name` will set the value
for the property's column head row. If no value is set, then the property name
is used. To leave the column head empty, you can set the name struct tag to `-`.

Additionally, you can designate a property to be skipped from being rendered entirely
by setting the `tabl` struct tag to `-`: ````tabl:"-"````. This is useful for storing
attributes that may be used during render calls for each row. Perhaps something like
a resource ID or row ID which may be injected into the table row attributes.

If your data type implements the method `(t T) TableRowAttributes() templ.Attributes`
then this function will be invoked when the row is is rendered and any attributes
returned will be set within the `<tr>` element for the row. Similarly, if the data
type implements the method `(t T) TableColumnAttributes(fieldName string)
templ.Attributes` then this function will be invoked when each rendered property
is being rendered and any returned attributes are applied to the `<td>` element.

### Example Data Type

```go
type Person struct {
  RowNumber  int    `name:"-"` // leave the column head empty
  FirstName  string `name:"First Name"` // column head value will be First Name
  LastName   string `name:"Last Name"` // column head value will be Last Name
  Age        int  // column head value will be Age
  Occupation string  // column head value will be Occupation
  Odd        bool `tabl:"-"` // this property and column is not rendered in the table
}

func (p Person) TableRowAttributes() templ.Attributes {
  class := "border border-collapse"
  if p.Odd {
    class += " border-slate-800"
  } else {
    class += " border-slate-600"
  }
  if p.Age > 40 {
    class += " bg-red-500/20"
  }
  return templ.Attributes{"class": class}
}

func (p Person) TableColumnAttributes(fieldName string) templ.Attributes {
  class := "border border-slate-800 border-collapse"
  if fieldName == "RowNumber" {
    class = class + " font-bold"
  }
  if fieldName == "Age" && p.Age > 40 {
    class = class + " text-red-500 font-bold"
  }
  return templ.Attributes{"class": class}
}
```

## Component Use

There are a few ways in which you can use the generated components depending
on situations and preference.

### Composition

Tabl generates several components for the Table, Table Head, Table Body,
and Rows. You can build your own component that uses each of these:

```templ
@PersonTable(nil) {
  <caption class="caption-bottom">
    Table 3.14: Persons guilty of a soggy bottom
  </caption>
  @PersonTableHead(nil, nil)
  @PersonTableBody(nil) {
    for _, person := range people {
      @PersonTableRow(person)
    }
  }
}
```

### Table Options

The generated {{Type}}TableOptions provides a more generic approach
to generating the table where some options properties can be set and
the data is fed in as a variadic parameter, `Table(data ...T)`:

```templ
@PersonTableOptions{
  TableAttributes: templ.Attributes{"class": "border border-black border-2"},
}.Table(
  Person{
    FirstName: "Rhonda",
    LastName: "Salana",
  },
  Person{
    FirstName: "Bo",
    LastName: "Sanchez",
  }
)
```

### Generated Aggregate Type

Tabl also generates an aggregate type for the collection of your
types with the `Components` suffix. This type has a
`Table({{Type}}TableOptions) templ.Component` method:

```templ
@PersonComponents{
  {
    FirstName: "Rhonda",
    LastName: "Salana",
  },
  {
    FirstName: "Bo",
    LastName: "Sanchez",
  }
}.Table(PersonTableOptions{
  TableAttributes: templ.Attributes{"class": "border border-black border-2"},
})
```

### Adding Attributes to Elements

Each generated component has its own way of applying attributes to the table
elements.

* The `<T>Table`  component simply accepts a `templ.Attributes`
parameter which is applied to the `<table>` element.
* The `<T>TableHead` component accepts:
  * a `templ.Attributes` paramater which is applied to the `<thead>` element.
  * a `func(fieldName string) templ.Attribute` param which is called on each
    render property column with the Go Field Name. The func should return
    attributes which will be applied to each `<th>` element.
* The `<T>TableBody` component accepts a `templ.Attributes` parameter which
is applied to the `<tbody>` element.
* The `<T>TableRow` attributes are set using Data Type methods, see
the `Data Types` section for more details.

#### Full Example

```templ
@PersonTable(templ.Attributes{"class": "border border-slate-800 border-collapse text-center rounded-md"}) {
  <caption class="caption-bottom">
    Table 3.14: Persons guilty of a soggy bottom
  </caption>
  @PersonTableHead(templ.Attributes{"class": "border border-slate-800 border-collapse bg-slate-700/30"}, func(string) templ.Attributes{
    return templ.Attributes{"class": "border border-slate-800 border-collapse"}
    })
  @PersonTableBody(templ.Attributes{"class": "border border-slate-800 border-collapse bg-gray-400/20"}) {
    for _, person := range people {
      @PersonTableRow(person)
    }
  }
}
```
