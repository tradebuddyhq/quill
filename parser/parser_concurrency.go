package parser

import (
	"quill/ast"
	"quill/lexer"
)

func (p *Parser) parseMount() *ast.MountStatement {
	line := p.current().Line
	p.advance() // consume "mount"
	comp := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_TO)
	selector := p.parseExpression()
	p.consumeNewline()
	return &ast.MountStatement{Component: comp.Value, Selector: selector, Line: line}
}

// --- Concurrency parsing ---

func (p *Parser) parseCancel() ast.Statement {
	line := p.current().Line
	p.advance() // consume "cancel"
	target := p.expect(lexer.TOKEN_IDENT)
	p.consumeNewline()
	return &ast.CancelStatement{Target: target.Value, Line: line}
}

func (p *Parser) parseSpawn() ast.Statement {
	line := p.current().Line
	p.advance() // consume "spawn"
	p.expect(lexer.TOKEN_TASK)
	nameTok := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()
	return &ast.SpawnStatement{Name: nameTok.Value, Body: body, Line: line}
}

func (p *Parser) parseParallel() ast.Statement {
	line := p.current().Line
	p.advance() // consume "parallel"

	// Check for "parallel settled:"
	isSettled := false
	if p.check(lexer.TOKEN_SETTLED) {
		isSettled = true
		p.advance() // consume "settled"
	}

	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()
	return &ast.ParallelBlock{Tasks: body, IsSettled: isSettled, Line: line}
}

func (p *Parser) parseRace() ast.Statement {
	line := p.current().Line
	p.advance() // consume "race"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	body := p.parseBlock()
	return &ast.RaceBlock{Tasks: body, Line: line}
}

func (p *Parser) parseChannel() ast.Statement {
	line := p.current().Line
	p.advance() // consume "channel"
	nameTok := p.expect(lexer.TOKEN_IDENT)
	var bufferSize ast.Expression
	if p.check(lexer.TOKEN_WITH) {
		p.advance() // consume "with"
		p.expect(lexer.TOKEN_BUFFER)
		bufferSize = p.parseExpression()
	}
	p.consumeNewline()
	return &ast.ChannelStatement{Name: nameTok.Value, BufferSize: bufferSize, Line: line}
}

func (p *Parser) parseSend() ast.Statement {
	line := p.current().Line
	p.advance() // consume "send"
	value := p.parseExpression()
	p.expect(lexer.TOKEN_TO)
	channelTok := p.expect(lexer.TOKEN_IDENT)
	p.consumeNewline()
	return &ast.SendStatement{Value: value, Channel: channelTok.Value, Line: line}
}

func (p *Parser) parseSelect() ast.Statement {
	line := p.current().Line
	p.advance() // consume "select"
	p.expect(lexer.TOKEN_COLON)
	p.expect(lexer.TOKEN_NEWLINE)
	p.expect(lexer.TOKEN_INDENT)

	var cases []ast.SelectCase
	var afterMs ast.Expression
	var afterBody []ast.Statement

	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
			break
		}
		if p.check(lexer.TOKEN_WHEN) {
			p.advance() // consume "when"
			p.expect(lexer.TOKEN_RECEIVE)
			p.expect(lexer.TOKEN_FROM)
			channelTok := p.expect(lexer.TOKEN_IDENT)
			p.expect(lexer.TOKEN_COLON)
			p.expect(lexer.TOKEN_NEWLINE)
			body := p.parseBlock()
			cases = append(cases, ast.SelectCase{Channel: channelTok.Value, Body: body})
		} else if p.check(lexer.TOKEN_AFTER) {
			p.advance() // consume "after"
			afterMs = p.parseExpression()
			p.expect(lexer.TOKEN_COLON)
			p.expect(lexer.TOKEN_NEWLINE)
			afterBody = p.parseBlock()
		} else {
			p.error("expected 'when' or 'after' in select block")
		}
	}
	if p.check(lexer.TOKEN_DEDENT) {
		p.advance()
	}

	return &ast.SelectStatement{Cases: cases, AfterMs: afterMs, AfterBody: afterBody, Line: line}
}

// --- Trait parsing ---

