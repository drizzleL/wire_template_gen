package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

// step1 根据import获取包名对应import的映射
// step2 根据var中的绑定关系找到,接口对应的实现映射
// step3 加载import中的包,找到对应的实体结构字段

var buf = bytes.NewBuffer(nil)

var output = "template/m_wire.go"
var templateFile = "template/m.go" // 映射模板文件

func main() {
	fs := token.NewFileSet()
	name := "template/m.go"
	parsedFile, err := parser.ParseFile(fs, name, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("parse package %s failed: %s", name, err)
	}

	loadPkgName(parsedFile)

	loadImportPkgDict(parsedFile)
	// log.Println("import pkg dict starting...")
	// printMap(importPkgDict)
	// log.Println("import pkg dict ending...")

	// write to buf, inf provider

	loadBind(parsedFile)
	// log.Println("bind dict starting...")
	// printMap(bindDict)
	// log.Println("bind dict ending...")

	writeInfProvider()

	loadImpl()

	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("format source faild:%s", err)
	}
	buf.Reset()

	b, err := imports.Process("template/m_wire.go", src, nil)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(output, b, 0644); err != nil {
		panic(err)
	}

	// for _, node := range parsedFile.Decls {
	// 	decl, ok := node.(*ast.GenDecl)
	// 	if !ok {
	// 		continue
	// 	}
	// 	if decl.Tok == token.VAR && strings.Contains(decl.Doc.Text(), "inf_mapping") {
	// 		dealWithInfVar(decl)
	// 	}
	// 	if decl.Tok == token.IMPORT {
	// 		// dealWithImport(decl)
	// 	}
	// }
}

func loadPkgName(parsedFile *ast.File) {
	fmt.Fprintf(buf, "package %s\n", parsedFile.Name.Name)
}

func writeInfProvider() {
	for infNameWithPkg, implNameWithPkg := range bindDict {
		infName := getNameByInf(infNameWithPkg, "")
		implName := getNameByInf(implNameWithPkg, "")
		tmp := buildInf(infName, infNameWithPkg, implName, implNameWithPkg)
		fmt.Fprintf(buf, tmp)
	}
}

// 根据import获取包名+importpath对应关系
var (
	importPkgDict = make(map[string]string)
)

func loadImportPkgDict(parsedFile *ast.File) {
	for _, node := range parsedFile.Decls {
		decl, ok := node.(*ast.GenDecl)
		if !ok {
			continue
		}
		if decl.Tok != token.IMPORT {
			continue
		}
		for _, spec := range decl.Specs {
			vspec := spec.(*ast.ImportSpec)
			importPath := vspec.Path.Value
			importPath = importPath[1 : len(importPath)-1]

			sepIdx := strings.LastIndexFunc(importPath, func(r rune) bool {
				return r == '/'
			})
			if sepIdx == -1 { // not found /, just package
				continue
			}
			pkgName := importPath[sepIdx+1:]
			if vspec.Name != nil {
				pkgName = vspec.Name.Name
			}
			importPkgDict[pkgName] = importPath
		}
		return
	}
	return
}

var (
	bindDict     = make(map[string]string) // key为结构体名
	implNameDict = make(map[string]bool)   // key为结构体名
)

func printMap(m map[string]string) {
	for k, v := range m {
		fmt.Println(k, ":", v)
	}
}

func loadBind(parsedFile *ast.File) {
	for _, node := range parsedFile.Decls {
		decl, ok := node.(*ast.GenDecl)
		if !ok {
			continue
		}
		if decl.Tok != token.VAR {
			continue
		}
		if !strings.Contains(decl.Doc.Text(), "inf_mapping") {
			continue
		}
		for _, spec := range decl.Specs {
			vspec := spec.(*ast.ValueSpec)
			// srvName := vspec.Names[0].Name
			infName := selectorToName(vspec.Type.(*ast.SelectorExpr))
			call := vspec.Values[0].(*ast.CallExpr) // 值,即实现,作为value
			value := call.Fun.(*ast.ParenExpr).X.(*ast.StarExpr).X.(*ast.SelectorExpr)
			implName := selectorToName(value)
			bindDict[infName] = implName
			implNameDict[implName] = true
		}
		return
	}
}

func loadImpl() {
	for pkgName, importPath := range importPkgDict {
		loadConfig := new(packages.Config)
		loadConfig.Mode = packages.NeedSyntax | packages.NeedDeps | packages.NeedFiles
		loadConfig.Fset = token.NewFileSet()
		pkgs, err := packages.Load(loadConfig, importPath)
		if err != nil {
			panic(err)
		}
		for _, f := range pkgs[0].Syntax {
			for _, decl := range f.Decls {
				gecl, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				if gecl.Tok != token.TYPE {
					continue
				}
				// 结构体是否存在与定义中,不存在可以跳过
				tspec := gecl.Specs[0].(*ast.TypeSpec)
				implNameWithPkg := pkgName + "." + tspec.Name.Name
				if !implNameDict[implNameWithPkg] { // 不存在的话跳过
					continue
				}
				styp, ok := tspec.Type.(*ast.StructType)
				if !ok { // 不是结构体则跳过
					continue
				}
				var fields []Field
				for _, field := range styp.Fields.List {
					typName := toTypeString(field.Type)
					if strings.Contains(typName, ".") { // 存在则说明外部引用包
					} else { // 不存在使用本身包名
						typName = pkgName + "." + typName
					}
					fields = append(fields, Field{
						Name: getNameByInf(typName, ""),
					})
				}
				implName := getNameByInf(implNameWithPkg, "")
				tmp := buildImpl(implName, implNameWithPkg, fields)
				fmt.Fprintf(buf, tmp)
			}
		}
	}
}

func toTypeString(exp ast.Expr) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), exp); err != nil {
		log.Fatalf("format ast.Expr failed:%s", err)
	}
	return buf.String()
}

var infToName = make(map[string]string)

func getNameByInf(infName string, backUp string) string {
	if name, ok := infToName[infName]; ok {
		return name
	}
	if backUp != "" {
		infToName[infName] = backUp
		return backUp
	}
	name := strings.Split(infName, ".")[1]
	infToName[infName] = name
	return name
}

func selectorToName(expr *ast.SelectorExpr) string {
	implName := expr.Sel.Name
	packageName := expr.X.(*ast.Ident).Name
	return packageName + "." + implName
}

// 参数1:接口名
// 参数2:包+接口名
// 参数3:实体名
// 参数4:包+实体名
const bindTmpl = `
var New%[1]s = wire.NewSet(
	wire.Bind(new(%[2]s), new(*%[4]s)),
	New%[3]s,
)
`

func buildInf(infName, infWithPkgName string, implName, implWithPkgName string) string {
	return fmt.Sprintf(bindTmpl, infName, infWithPkgName, implName, implWithPkgName)
}

// 参数1:实体名
// 参数2:包+实体名
const implConstructorTmplNoFields = `
func New%[1]s() *%[2]s {
	panic(wire.Build(
		wire.Struct(new(%[2]s), "*"),
	))
}`

// 参数1:实体名
// 参数2:包+实体名
// 参数3:字段provider
const implConstructorTmpl = `
func New%[1]s() *%[2]s {
	panic(wire.Build(
		wire.Struct(new(%[2]s), "*"),%[3]s
	))
}`

type Field struct {
	Name string
}

func buildImpl(implName string, implWithPkgName string, fields []Field) string {
	if len(fields) == 0 {
		return fmt.Sprintf(implConstructorTmplNoFields, implName, implWithPkgName)
	}
	var fieldsStr string
	for _, field := range fields {
		fieldsStr += "\n" + getProviderName(field.Name) + ","
	}
	return fmt.Sprintf(implConstructorTmpl, implName, implWithPkgName, fieldsStr)
}

func getProviderName(name string) string {
	return fmt.Sprintf("New" + name)
}
