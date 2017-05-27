package main

import (
	"os"
	"github.com/yhat/scrape"
	"strings"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Verifier func(string) bool
type TypeEvaluator func(Verifier, Verifier) bool
type IndexEntry struct {
	name string
	elementType ElementType
	path string
}

var ALL_ELEMENT_TYPES = []ElementType{
	Class,
	Method,
	Field,
	Constructor,
	Interface,
	Exception,
	Error,
	Enum,
	Trait,
	Notation,
	Package,
}

var ELEMENT_TYPE_TO_TYPE_EVALUATORS = map[ElementType][]TypeEvaluator{
	Class: 		NewTypeEvaluators(isClass),
	Method: 	NewTypeEvaluators(isStaticMethod, isMethod),
	Field:	 	NewTypeEvaluators(isStaticField, isField),
	Constructor: 	NewTypeEvaluators(isConstructor),
	Interface: 	NewTypeEvaluators(isInterface),
	Exception: 	NewTypeEvaluators(isException),
	Error: 		NewTypeEvaluators(isError),
	Enum: 		NewTypeEvaluators(isEnum),
	Trait: 		NewTypeEvaluators(isTrait),
	Notation: 	NewTypeEvaluators(isNotation),
	Package: 	NewTypeEvaluators(isPackage),
}

func parseIndex(indexFilePath string, entryHandler func(IndexEntry)) {

	log.Info("Indexing from file", "file", indexFilePath)

	indexed := 0
	file, err := os.OpenFile(indexFilePath, os.O_RDONLY, 0666)

	if err != nil {
		log.Error("Unable to open file", "file", indexFilePath, "error", err)
		return
	}

	root, err := html.Parse(file)

	if err != nil {
		log.Error("Unable to parse index", "file", file, "error", err)
		return
	}

	anchorTags := scrape.FindAll(root, scrape.ByTag(atom.A))

	for _, tag := range anchorTags {
		var parentTag = tag.Parent

		if parentTag.FirstChild != tag {
			continue
		}

		isParentSpan := parentTag.DataAtom == atom.Span
		isParentCode := parentTag.DataAtom == atom.Code
		isParentItalic := parentTag.DataAtom == atom.I
		isParentBold := parentTag.DataAtom == atom.B

		if isParentSpan || isParentCode || isParentItalic || isParentBold {
			parentTag = parentTag.Parent
			if parentTag.FirstChild != tag.Parent {
				continue
			}
		}

		if parentTag.DataAtom != atom.Dt {
			continue
		}

		text := scrape.Text(parentTag)
		var tagType ElementType = NotFound
		var dtClassName = scrape.Attr(parentTag, "class")

		lowercaseText := strings.ToLower(text)

		textContainsInsensitive := func(s string) bool {
			return strings.Contains(lowercaseText, s)
		}

		dtClassNameHasSuffix := func(s string) bool {
			return strings.HasSuffix(dtClassName, s)
		}

		tagTypeDetermined := false

		for _, elementType := range ALL_ELEMENT_TYPES {

			typeEvaluators := ELEMENT_TYPE_TO_TYPE_EVALUATORS[elementType]

			for _, evaluator := range typeEvaluators {

				if evaluator(textContainsInsensitive, dtClassNameHasSuffix) {
					tagType = elementType
					tagTypeDetermined = true
					break
				}
			}

			if tagTypeDetermined {
				break
			}
		}

		if tagType == NotFound {
			log.Warn("Warning: could not determine type", "text", text, "dtClassName", dtClassName)
			continue
		}

		name := scrape.Text(tag)
		path := scrape.Attr(tag, "href")

		entryHandler(IndexEntry{name: name, elementType: tagType, path: path})

		indexed++
	}

	log.Info("Indexed", "count", indexed)
}

func NewTypeEvaluators(a TypeEvaluator, others ...TypeEvaluator) []TypeEvaluator {

	typeEvaluators := make([]TypeEvaluator, 1 + len(others))

	typeEvaluators[0] = a
	for i, typeEvaluator := range others {
		typeEvaluators[i + 1] = typeEvaluator
	}

	return typeEvaluators
}

func isClass(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("class in") || textContainsInsensitive("- class") || dtClassNameHasSuffix("class")
}

func isStaticMethod(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("static method in") || dtClassNameHasSuffix("method")
}

func isStaticField(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("static variable in") || textContainsInsensitive("field in") || dtClassNameHasSuffix("field")
}

func isConstructor(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("constructor") || dtClassNameHasSuffix("constructor")
}

func isMethod(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("method in")
}

func isField(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("variable in")
}

func isInterface(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("interface in") || textContainsInsensitive("- interface") || dtClassNameHasSuffix("interface")
}

func isException(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("exception in") || textContainsInsensitive("- exception") || dtClassNameHasSuffix("exception")
}

func isError(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("error in") || textContainsInsensitive("- error") || dtClassNameHasSuffix("error")
}

func isEnum(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("enum in") || textContainsInsensitive("- enum") || dtClassNameHasSuffix("enum")
}

func isTrait(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("trait in")
}

func isNotation(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("annotation type") || dtClassNameHasSuffix("annotation")
}

func isPackage(textContainsInsensitive, dtClassNameHasSuffix Verifier) bool {
	return textContainsInsensitive("package") || dtClassNameHasSuffix("package")
}
