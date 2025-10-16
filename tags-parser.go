// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bin

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// TagNodeType 定义标签节点类型
type TagNodeType int

const (
	TagTypeModifier  TagNodeType = iota // 修饰符节点
	TagTypeValue                        // 值节点
	TagTypeParameter                    // 参数节点
)

// TagNode 表示标签的抽象语法树节点
type TagNode struct {
	Type     TagNodeType
	Name     string      // 节点名称 (如 "hidden_prefix", "fixed_size")
	Value    interface{} // 叶子节点的值
	Children []*TagNode  // 子节点
}

// String 返回节点的字符串表示
func (node *TagNode) String() string {
	return node.stringWithIndent(0)
}

// FindChildByName 按照子节点的名字查找子节点
func (node *TagNode) FindChildByName(name string) *TagNode {
	for _, child := range node.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

func (node *TagNode) stringWithIndent(indent int) string {
	prefix := strings.Repeat("  ", indent)

	if node.Type == TagTypeValue || node.Type == TagTypeParameter {
		return fmt.Sprintf("%s- %s: %v\n", prefix, node.Name, node.Value)
	}

	result := fmt.Sprintf("%s- %s\n", prefix, node.Name)
	for _, child := range node.Children {
		result += child.stringWithIndent(indent + 1)
	}
	return result
}

type fieldTag struct {
	SizeOf          string
	Skip            bool
	Order           binary.ByteOrder
	Option          bool
	COption         bool
	BinaryExtension bool

	IsBorshEnum bool

	NestedTag *TagNode // 新增:嵌套标签的 AST
}

func isIn(s string, candidates ...string) bool {
	for _, c := range candidates {
		if s == c {
			return true
		}
	}
	return false
}

func parseFieldTag(tag reflect.StructTag) *fieldTag {
	t := &fieldTag{
		Order: defaultByteOrder,
	}
	tagStr := tag.Get("bin")
	// 尝试解析嵌套标签
	if strings.Contains(tagStr, "<") {
		nestedTag, err := parseNestedTag(tagStr)
		if err == nil {
			t.NestedTag = nestedTag
			return t
		}
		// 如果解析失败，回退到简单标签解析
	}
	for _, s := range strings.Split(tagStr, " ") {
		if strings.HasPrefix(s, "sizeof=") {
			tmp := strings.SplitN(s, "=", 2)
			t.SizeOf = tmp[1]
		} else if s == "big" {
			t.Order = binary.BigEndian
		} else if s == "little" {
			t.Order = binary.LittleEndian
		} else if isIn(s, "optional", "option") {
			t.Option = true
		} else if isIn(s, "coption") {
			t.COption = true
		} else if s == "binary_extension" {
			t.BinaryExtension = true
		} else if isIn(s, "-", "skip") {
			t.Skip = true
		} else if isIn(s, "enum") {
			t.IsBorshEnum = true
		}
	}

	// TODO: parse other borsh tags
	if strings.TrimSpace(tag.Get("borsh_skip")) == "true" {
		t.Skip = true
	}
	if strings.TrimSpace(tag.Get("borsh_enum")) == "true" {
		t.IsBorshEnum = true
	}
	return t
}

// parseNestedTag 解析嵌套的标签字符串
func parseNestedTag(tag string) (*TagNode, error) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return nil, fmt.Errorf("empty tag")
	}

	// 查找第一个 '<' 的位置
	openBracket := strings.Index(tag, "<")

	if openBracket == -1 {
		// 叶子节点 - 尝试解析为数字或字符串
		if val, err := strconv.Atoi(tag); err == nil {
			return &TagNode{
				Type:  TagTypeParameter,
				Name:  "number",
				Value: val,
			}, nil
		}
		return &TagNode{
			Type:  TagTypeValue,
			Name:  tag,
			Value: tag,
		}, nil
	}

	// 提取修饰符名称
	modifierName := strings.TrimSpace(tag[:openBracket])

	// 查找匹配的 '>'
	closeBracket := findMatchingBracket(tag, openBracket)
	if closeBracket == -1 {
		return nil, fmt.Errorf("unmatched '<' in tag: %s", tag)
	}

	// 提取括号内的内容
	innerContent := tag[openBracket+1 : closeBracket]

	// 解析子节点
	children, err := parseChildren(innerContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse children of %s: %w", modifierName, err)
	}

	return &TagNode{
		Type:     TagTypeModifier,
		Name:     modifierName,
		Children: children,
	}, nil
}

// findMatchingBracket 查找匹配的右括号
func findMatchingBracket(s string, openPos int) int {
	depth := 1
	for i := openPos + 1; i < len(s); i++ {
		switch s[i] {
		case '<':
			depth++
		case '>':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// parseChildren 解析逗号分隔的子节点
func parseChildren(content string) ([]*TagNode, error) {
	if content == "" {
		return nil, nil
	}

	var children []*TagNode
	var current strings.Builder
	depth := 0

	for i := 0; i < len(content); i++ {
		ch := content[i]
		switch ch {
		case '<':
			depth++
			current.WriteByte(ch)
		case '>':
			depth--
			current.WriteByte(ch)
		case ',':
			if depth == 0 {
				// 遇到顶层逗号,解析当前累积的内容
				child, err := parseNestedTag(current.String())
				if err != nil {
					return nil, err
				}
				children = append(children, child)
				current.Reset()
			} else {
				current.WriteByte(ch)
			}
		default:
			current.WriteByte(ch)
		}
	}

	// 解析最后一个子节点
	if current.Len() > 0 {
		child, err := parseNestedTag(current.String())
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}

	return children, nil
}
