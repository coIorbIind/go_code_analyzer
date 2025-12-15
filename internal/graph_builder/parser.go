package graph_builder

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type Config struct {
	IncludePkgs map[string]bool
	ExcludePkgs map[string]bool
}

type analysisContext struct {
	pkg          *packages.Package
	currentScope string
	varBindings  map[string]string // имя переменной -> полное имя замыкания
}

func newAnalysisContext(pkg *packages.Package, scope string) *analysisContext {
	return &analysisContext{
		pkg:          pkg,
		currentScope: scope,
		varBindings:  make(map[string]string),
	}
}

func (ctx *analysisContext) getScopedName(varName string) string {
	if ctx.currentScope != "" {
		return fmt.Sprintf("%s.%s", ctx.currentScope, varName)
	}
	return fmt.Sprintf("%s.%s", ctx.pkg.Name, varName)
}

func (ctx *analysisContext) bindVariable(varName, closureName string) {
	ctx.varBindings[varName] = closureName
}

func (ctx *analysisContext) resolveVariable(varName string) string {
	if name, exists := ctx.varBindings[varName]; exists {
		return name
	}
	return fmt.Sprintf("%s.%s", ctx.pkg.Name, varName)
}

func AnalyzeProject(projectDir string, g *Graph, cfg *Config) error {
	fset := token.NewFileSet()

	config := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Dir:  projectDir,
		Fset: fset,
	}

	pkgs, err := packages.Load(config, "./...")
	if err != nil {
		return fmt.Errorf("packages.Load failed: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("errors found during package loading")
	}

	for _, pkg := range pkgs {
		if pkg.Types == nil || len(pkg.Syntax) == 0 {
			continue
		}

		pkgName := pkg.Name

		if len(cfg.IncludePkgs) > 0 && !cfg.IncludePkgs[pkgName] {
			continue
		}
		if cfg.ExcludePkgs[pkgName] {
			continue
		}

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.FuncDecl:
					funcName := getFunctionName(node, pkg)
					g.AddNode(funcName)
					ctx := newAnalysisContext(pkg, funcName)
					if node.Body != nil {
						processFuncBody(node.Body, g, pkg, ctx)
					}
					return false
				}
				return true
			})
		}
	}

	return nil
}

func processFuncBody(body *ast.BlockStmt, g *Graph, pkg *packages.Package, ctx *analysisContext) {
	if body == nil {
		return
	}

	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			callName := getCallName(node, pkg, ctx)
			if callName == "" {
				return true
			}

			caller := ctx.currentScope
			if caller == "" {
				return true
			}

			g.AddNode(callName)
			g.AddEdge(caller, callName)

		case *ast.AssignStmt:
			processAssignments(node, g, pkg, ctx)

		case *ast.FuncLit:
			// Inner closures
			return false
		}
		return true
	})
}

func processAssignments(assign *ast.AssignStmt, g *Graph, pkg *packages.Package, ctx *analysisContext) {
	for i, rhs := range assign.Rhs {
		if fn, ok := rhs.(*ast.FuncLit); ok && i < len(assign.Lhs) {
			if ident, ok := assign.Lhs[i].(*ast.Ident); ok {
				closureName := ctx.getScopedName(ident.Name)
				ctx.bindVariable(ident.Name, closureName)
				g.AddNode(closureName)

				newCtx := newAnalysisContext(pkg, closureName)

				for k, v := range ctx.varBindings {
					newCtx.bindVariable(k, v)
				}

				if fn.Body != nil {
					processFuncBody(fn.Body, g, pkg, newCtx)
				}
			}
		}
	}
}

func getCallName(call *ast.CallExpr, pkg *packages.Package, ctx *analysisContext) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		// Closure call
		if resolvedName := ctx.resolveVariable(fn.Name); resolvedName != fmt.Sprintf("%s.%s", ctx.pkg.Name, fn.Name) {
			return resolvedName
		}

		if pkg.TypesInfo != nil {
			if obj := pkg.TypesInfo.Uses[fn]; obj != nil {
				if obj.Pkg() == nil {
					return fn.Name
				}
			}
		}
		return fmt.Sprintf("%s.%s", pkg.Name, fn.Name)

	case *ast.SelectorExpr:
		if pkg.TypesInfo != nil {
			if selection := pkg.TypesInfo.Selections[fn]; selection != nil {
				recvType := selection.Recv()
				methodName := selection.Obj().Name()

				typeName := getTypeName(recvType)
				if typeName != "" {
					return fmt.Sprintf("%s.%s", typeName, methodName)
				}
			}

			if ident, ok := fn.X.(*ast.Ident); ok {
				if obj := pkg.TypesInfo.Uses[ident]; obj != nil {
					if typeName := getTypeName(obj.Type()); typeName != "" {
						return fmt.Sprintf("%s.%s", typeName, fn.Sel.Name)
					}
					if obj.Pkg() != nil {
						return fmt.Sprintf("%s.%s", obj.Pkg().Name(), fn.Sel.Name)
					}
				}
			}
		}

		if ident, ok := fn.X.(*ast.Ident); ok {
			if resolved := ctx.resolveVariable(ident.Name); resolved != fmt.Sprintf("%s.%s", ctx.pkg.Name, ident.Name) {
				return fmt.Sprintf("%s.%s", resolved, fn.Sel.Name)
			}

			return fmt.Sprintf("%s.%s", ident.Name, fn.Sel.Name)
		}
		return fn.Sel.Name
	}
	return ""
}

func getTypeName(typ types.Type) string {
	if typ == nil {
		return ""
	}

	switch t := typ.(type) {
	case *types.Pointer:
		return getTypeName(t.Elem())
	case *types.Named:
		if t.Obj() != nil && t.Obj().Pkg() != nil {
			pkgName := t.Obj().Pkg().Name()
			typeName := t.Obj().Name()

			if pkgName == "sync" {
				return fmt.Sprintf("%s.%s", pkgName, typeName)
			}
			return fmt.Sprintf("%s.%s", pkgName, typeName)
		} else if t.Obj() != nil {
			return t.Obj().Name()
		}
	case *types.Struct:
		return "struct"
	}

	return ""
}

func getFunctionName(fn *ast.FuncDecl, pkg *packages.Package) string {
	funcName := fn.Name.Name
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recv := fn.Recv.List[0]
		switch t := recv.Type.(type) {
		case *ast.Ident:
			return fmt.Sprintf("%s.%s.%s", pkg.Name, t.Name, funcName)
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				return fmt.Sprintf("%s.%s.%s", pkg.Name, ident.Name, funcName)
			}
		}
	}
	return fmt.Sprintf("%s.%s", pkg.Name, funcName)
}
