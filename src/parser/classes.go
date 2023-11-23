package parser

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
	"github.com/vyPal/CaffeineC/compiler"
)

func (p *Parser) parseClassDefinition() *compiler.Class {
	p.Pos++ // "class"
	name := p.Tokens[p.Pos].Value
	p.Pos++ // name
	p.Pos++ // "{"
	var fieldCount int = 0
	var fields []compiler.Field
	var methods []compiler.Method
	var constructor compiler.Method
	for p.Tokens[p.Pos].Type != "PUNCT" || p.Tokens[p.Pos].Value != "}" {
		if p.Tokens[p.Pos].Value == "func" {
			methods = append(methods, p.parseMethod())
		} else if p.Tokens[p.Pos].Value == "private" {
			if p.Tokens[p.Pos+1].Value == "func" {
				methods = append(methods, p.parseMethod())
			} else {
				fields = append(fields, p.parseField(fieldCount))
				fieldCount++
			}
		} else if p.Tokens[p.Pos].Value == name {
			constructor = p.parseClassConstructor()
		} else {
			fields = append(fields, p.parseField(fieldCount))
			fieldCount++
		}
	}
	p.Pos++ // "}"
	return &compiler.Class{Name: name, Fields: fields, Methods: methods, Constructor: constructor}
}

func (p *Parser) parseClassConstructor() compiler.Method {
	name := p.Tokens[p.Pos].Value
	p.Pos++ // name
	p.Pos++ // "("
	// Parse the parameters
	var params []*ir.Param
	for p.Tokens[p.Pos].Type != "PUNCT" || p.Tokens[p.Pos].Value != ")" {
		paramName := p.Tokens[p.Pos].Value
		p.Pos++ // name
		p.Pos++ // ":"
		paramType := p.Tokens[p.Pos].Value
		p.Pos++ // type
		switch paramType {
		case "int":
			params = append(params, ir.NewParam(paramName, types.I64))
		case "string":
			params = append(params, ir.NewParam(paramName, &types.ArrayType{ElemType: types.I8}))
		case "float64":
			params = append(params, ir.NewParam(paramName, types.Double))
		case "bool":
			params = append(params, ir.NewParam(paramName, types.I1))
		case "duration":
			params = append(params, ir.NewParam(paramName, types.I64))
		default:
			panic(fmt.Sprintf("Unknown type %s", paramType))
		}
		if p.Tokens[p.Pos].Type == "PUNCT" && p.Tokens[p.Pos].Value == "," {
			p.Pos++ // ","
		}
	}
	p.Pos++ // ")"
	p.Pos++
	fmt.Println("Start of function", name)
	var body []compiler.Stmt
	for p.Tokens[p.Pos].Type != "PUNCT" || p.Tokens[p.Pos].Value != "}" {
		token := p.Tokens[p.Pos]
		switch token.Type {
		case "IDENT":
			if token.Value == "var" {
				body = append(body, p.parseVarDecl()...)
			} else if token.Value == "return" {
				p.Pos++ // "return"
				value := p.parseExpression()
				fmt.Println("Return", value)
				body = append(body, &compiler.SRet{Val: value})
			} else if token.Value == "print" {
				body = append(body, p.parsePrint())
			} else if token.Value == "sleep" {
				body = append(body, p.parseSleep())
			} else if token.Value == "func" {
				body = append(body, p.parseFunctionDeclaration())
			} else {
				body = append(body, p.parseAssignment())
			}
		default:
			fmt.Println("[W]", token.Location, "Unexpected token:", token.Value)
			p.Pos++
		}
	}
	fmt.Println("End of function", name)
	p.Pos++ // "}"
	return compiler.Method{Name: name, Params: params, ReturnType: types.Void, Body: body, Private: false}
}

func (p *Parser) parseMethod() compiler.Method {
	var private bool = false
	if p.Tokens[p.Pos].Value == "private" {
		private = true
		p.Pos++ // "private"
	}
	funcDecl := p.parseFunctionDeclaration().(*compiler.SFuncDecl)
	return compiler.Method{Name: funcDecl.Name, Params: funcDecl.Args, ReturnType: funcDecl.ReturnType, Body: funcDecl.Body, Private: private}
}

func (p *Parser) parseField(index int) compiler.Field {
	name := p.Tokens[p.Pos].Value
	var private bool = false
	if name == "private" {
		private = true
		p.Pos++ // "private"
		name = p.Tokens[p.Pos].Value
	}
	p.Pos++ // name
	p.Pos++ // ":"
	var fieldType types.Type
	if p.Tokens[p.Pos].Type == "IDENT" {
		typ := p.Tokens[p.Pos].Value
		switch typ {
		case "int":
			fieldType = types.I64
		case "string":
			fieldType = types.NewPointer(types.I8)
		case "float64":
			fieldType = types.Double
		case "bool":
			fieldType = types.I1
		case "duration":
			fieldType = types.I64
		default:
			panic(fmt.Errorf("Unknown type: %s", typ))
		}
		p.Pos++ // type
	} else {
		panic("Field type not specified")
	}
	p.Pos++ // ";"
	return compiler.Field{Name: name, Type: fieldType, Private: private, Index: index}
}
