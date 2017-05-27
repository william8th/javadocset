package main

type ElementType int

const (
	NotFound ElementType = iota
	Class
	Method
	Field
	Constructor
	Interface
	Exception
	Error
	Enum
	Trait
	Notation
	Package
)

var VALUES = map[ElementType]string{
	Class: "Class",
	Method: "Method",
	Field: "Field",
	Constructor: "Constructor",
	Interface: "Interface",
	Exception: "Exception",
	Error: "Error",
	Enum: "Enum",
	Trait: "Trait",
	Notation: "Notation",
	Package: "Package",
}

func (e *ElementType) value() string {
	return VALUES[*e]
}
