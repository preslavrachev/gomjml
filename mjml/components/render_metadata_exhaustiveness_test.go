package components

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// indirectlyHandledFontTypes lists component types that resolve
// constants.MJMLFontFamily but are deliberately absent from the
// collectMetadataFromComponent type switch because a sibling case reaches
// them through its own helper instead (see render_metadata.go):
//   - MJAccordionTextComponent / MJAccordionTitleComponent: read via
//     MJAccordionElementComponent.renderedTitleAndText's explicitFontFamily.
//   - MJSocialElementComponent: read via collectSocialElementFont, called
//     from the MJSocialComponent case.
var indirectlyHandledFontTypes = map[string]bool{
	"MJAccordionTextComponent":  true,
	"MJAccordionTitleComponent": true,
	"MJSocialElementComponent":  true,
}

// TestRenderMetadataSwitchIsExhaustive guards against the failure mode
// documented on collectMetadataFromComponent: a new or edited component
// that resolves constants.MJMLFontFamily but is never added as a case (or
// to indirectlyHandledFontTypes above) would silently drop its font from
// the document head, with no compiler error, since Go does not check
// type-switch exhaustiveness over an interface. This test parses the
// package source directly so it fails as soon as such a component is
// introduced, rather than only when someone happens to write a font
// integration test for it.
func TestRenderMetadataSwitchIsExhaustive(t *testing.T) {
	fset := token.NewFileSet()

	touchesFontFamily := findTypesTouchingFontFamily(t, fset)
	handledInSwitch := findSwitchCaseTypes(t, fset, "render_metadata.go", "collectMetadataFromComponent")

	var missing []string
	for typ := range touchesFontFamily {
		if handledInSwitch[typ] || indirectlyHandledFontTypes[typ] {
			continue
		}
		missing = append(missing, typ)
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("components resolve constants.MJMLFontFamily but are not handled by "+
			"collectMetadataFromComponent (add a case in render_metadata.go, or add to "+
			"indirectlyHandledFontTypes if a sibling case already reaches them): %v", missing)
	}

	var staleAllowlist []string
	for typ := range indirectlyHandledFontTypes {
		if !touchesFontFamily[typ] {
			staleAllowlist = append(staleAllowlist, typ)
		}
	}
	sort.Strings(staleAllowlist)
	if len(staleAllowlist) > 0 {
		t.Errorf("indirectlyHandledFontTypes lists types that no longer resolve "+
			"constants.MJMLFontFamily; remove them: %v", staleAllowlist)
	}
}

// findTypesTouchingFontFamily scans every non-test .go file in this package
// for methods whose body references constants.MJMLFontFamily, returning the
// set of *MJXxxComponent receiver type names.
func findTypesTouchingFontFamily(t *testing.T, fset *token.FileSet) map[string]bool {
	t.Helper()
	result := make(map[string]bool)

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading package dir: %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", name, err)
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 || fn.Body == nil {
				continue
			}
			recvType := receiverTypeName(fn.Recv.List[0].Type)
			if !strings.HasPrefix(recvType, "MJ") || !strings.HasSuffix(recvType, "Component") {
				continue
			}
			if referencesFontFamilyConstant(fn.Body) {
				result[recvType] = true
			}
		}
	}

	return result
}

func receiverTypeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func referencesFontFamilyConstant(node ast.Node) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if ok && pkg.Name == "constants" && sel.Sel.Name == "MJMLFontFamily" {
			found = true
			return false
		}
		return true
	})
	return found
}

// findSwitchCaseTypes parses fileName and returns the *MJXxxComponent type
// names listed as case labels in the type switch inside function funcName.
func findSwitchCaseTypes(t *testing.T, fset *token.FileSet, fileName, funcName string) map[string]bool {
	t.Helper()
	result := make(map[string]bool)

	file, err := parser.ParseFile(fset, filepath.Join(".", fileName), nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", fileName, err)
	}

	var target *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == funcName {
			target = fn
			break
		}
	}
	if target == nil {
		t.Fatalf("function %s not found in %s", funcName, fileName)
	}

	ast.Inspect(target, func(n ast.Node) bool {
		sw, ok := n.(*ast.TypeSwitchStmt)
		if !ok {
			return true
		}
		for _, stmt := range sw.Body.List {
			clause, ok := stmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			for _, caseExpr := range clause.List {
				if typeName := receiverTypeName(caseExpr); typeName != "" {
					result[typeName] = true
				}
			}
		}
		return false
	})

	return result
}
