/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser2"
)

func TestConstantSizedType_String(t *testing.T) {

	t.Parallel()

	ty := &ConstantSizedType{
		Type: &VariableSizedType{Type: IntType},
		Size: 2,
	}

	assert.Equal(t,
		"[[Int]; 2]",
		ty.String(),
	)
}

func TestConstantSizedType_String_OfFunctionType(t *testing.T) {

	t.Parallel()

	ty := &ConstantSizedType{
		Type: &FunctionType{
			Parameters: []*Parameter{
				{
					TypeAnnotation: NewTypeAnnotation(Int8Type),
				},
			},
			ReturnTypeAnnotation: NewTypeAnnotation(
				Int16Type,
			),
		},
		Size: 2,
	}

	assert.Equal(t,
		"[((Int8): Int16); 2]",
		ty.String(),
	)
}

func TestVariableSizedType_String(t *testing.T) {

	t.Parallel()

	ty := &VariableSizedType{
		Type: &ConstantSizedType{
			Type: IntType,
			Size: 2,
		},
	}

	assert.Equal(t,
		"[[Int; 2]]",
		ty.String(),
	)
}

func TestVariableSizedType_String_OfFunctionType(t *testing.T) {

	t.Parallel()

	ty := &VariableSizedType{
		Type: &FunctionType{
			Parameters: []*Parameter{
				{
					TypeAnnotation: NewTypeAnnotation(Int8Type),
				},
			},
			ReturnTypeAnnotation: NewTypeAnnotation(
				Int16Type,
			),
		},
	}

	assert.Equal(t,
		"[((Int8): Int16)]",
		ty.String(),
	)
}

func TestIsResourceType_AnyStructNestedInArray(t *testing.T) {

	t.Parallel()

	ty := &VariableSizedType{
		Type: AnyStructType,
	}

	assert.False(t, ty.IsResourceType())
}

func TestIsResourceType_AnyResourceNestedInArray(t *testing.T) {

	t.Parallel()

	ty := &VariableSizedType{
		Type: AnyResourceType,
	}

	assert.True(t, ty.IsResourceType())
}

func TestIsResourceType_ResourceNestedInArray(t *testing.T) {

	t.Parallel()

	ty := &VariableSizedType{
		Type: &CompositeType{
			Kind: common.CompositeKindResource,
		},
	}

	assert.True(t, ty.IsResourceType())
}

func TestIsResourceType_ResourceNestedInDictionary(t *testing.T) {

	t.Parallel()

	ty := &DictionaryType{
		KeyType: StringType,
		ValueType: &VariableSizedType{
			Type: &CompositeType{
				Kind: common.CompositeKindResource,
			},
		},
	}

	assert.True(t, ty.IsResourceType())
}

func TestIsResourceType_StructNestedInDictionary(t *testing.T) {

	t.Parallel()

	ty := &DictionaryType{
		KeyType: StringType,
		ValueType: &VariableSizedType{
			Type: &CompositeType{
				Kind: common.CompositeKindStructure,
			},
		},
	}

	assert.False(t, ty.IsResourceType())
}

func TestRestrictedType_StringAndID(t *testing.T) {

	t.Parallel()

	t.Run("base type and restriction", func(t *testing.T) {

		t.Parallel()

		interfaceType := &InterfaceType{
			CompositeKind: common.CompositeKindResource,
			Identifier:    "I",
			Location:      common.StringLocation("b"),
		}

		ty := &RestrictedType{
			Type: &CompositeType{
				Kind:       common.CompositeKindResource,
				Identifier: "R",
				Location:   common.StringLocation("a"),
			},
			Restrictions: []*InterfaceType{interfaceType},
		}

		assert.Equal(t,
			"R{I}",
			ty.String(),
		)

		assert.Equal(t,
			TypeID("S.a.R{S.b.I}"),
			ty.ID(),
		)
	})

	t.Run("base type and restrictions", func(t *testing.T) {

		t.Parallel()

		i1 := &InterfaceType{
			CompositeKind: common.CompositeKindResource,
			Identifier:    "I1",
			Location:      common.StringLocation("b"),
		}

		i2 := &InterfaceType{
			CompositeKind: common.CompositeKindResource,
			Identifier:    "I2",
			Location:      common.StringLocation("c"),
		}

		ty := &RestrictedType{
			Type: &CompositeType{
				Kind:       common.CompositeKindResource,
				Identifier: "R",
				Location:   common.StringLocation("a"),
			},
			Restrictions: []*InterfaceType{i1, i2},
		}

		assert.Equal(t,
			ty.String(),
			"R{I1, I2}",
		)

		assert.Equal(t,
			TypeID("S.a.R{S.b.I1,S.c.I2}"),
			ty.ID(),
		)
	})

	t.Run("no restrictions", func(t *testing.T) {

		t.Parallel()

		ty := &RestrictedType{
			Type: &CompositeType{
				Kind:       common.CompositeKindResource,
				Identifier: "R",
				Location:   common.StringLocation("a"),
			},
		}

		assert.Equal(t,
			"R{}",
			ty.String(),
		)

		assert.Equal(t,
			TypeID("S.a.R{}"),
			ty.ID(),
		)
	})
}

func TestRestrictedType_Equals(t *testing.T) {

	t.Parallel()

	t.Run("same base type and more restrictions", func(t *testing.T) {

		t.Parallel()

		i1 := &InterfaceType{
			CompositeKind: common.CompositeKindResource,
			Identifier:    "I1",
			Location:      common.StringLocation("b"),
		}

		i2 := &InterfaceType{
			CompositeKind: common.CompositeKindResource,
			Identifier:    "I2",
			Location:      common.StringLocation("b"),
		}

		a := &RestrictedType{
			Type: &CompositeType{
				Kind:       common.CompositeKindResource,
				Identifier: "R",
				Location:   common.StringLocation("a"),
			},
			Restrictions: []*InterfaceType{i1},
		}

		b := &RestrictedType{
			Type: &CompositeType{
				Kind:       common.CompositeKindResource,
				Identifier: "R",
				Location:   common.StringLocation("a"),
			},
			Restrictions: []*InterfaceType{i1, i2},
		}

		assert.False(t, a.Equal(b))
	})

	t.Run("same base type and fewer restrictions", func(t *testing.T) {

		t.Parallel()

		i1 := &InterfaceType{
			CompositeKind: common.CompositeKindResource,
			Identifier:    "I1",
			Location:      common.StringLocation("b"),
		}

		i2 := &InterfaceType{
			CompositeKind: common.CompositeKindResource,
			Identifier:    "I2",
			Location:      common.StringLocation("b"),
		}

		a := &RestrictedType{
			Type: &CompositeType{
				Kind:       common.CompositeKindResource,
				Identifier: "R",
				Location:   common.StringLocation("a"),
			},
			Restrictions: []*InterfaceType{i1, i2},
		}

		b := &RestrictedType{
			Type: &CompositeType{
				Kind:       common.CompositeKindResource,
				Identifier: "R",
				Location:   common.StringLocation("a"),
			},
			Restrictions: []*InterfaceType{i1},
		}

		assert.False(t, a.Equal(b))
	})

	t.Run("same base type and same restrictions", func(t *testing.T) {

		t.Parallel()

		i1 := &InterfaceType{
			CompositeKind: common.CompositeKindResource,
			Identifier:    "I1",
			Location:      common.StringLocation("b"),
		}

		i2 := &InterfaceType{
			CompositeKind: common.CompositeKindResource,
			Identifier:    "I2",
			Location:      common.StringLocation("b"),
		}

		a := &RestrictedType{
			Type: &CompositeType{
				Kind:       common.CompositeKindResource,
				Identifier: "R",
				Location:   common.StringLocation("a"),
			},
			Restrictions: []*InterfaceType{i1, i2},
		}

		b := &RestrictedType{
			Type: &CompositeType{
				Kind:       common.CompositeKindResource,
				Identifier: "R",
				Location:   common.StringLocation("a"),
			},
			Restrictions: []*InterfaceType{i1, i2},
		}

		assert.True(t, a.Equal(b))
	})

	t.Run("different base type and same restrictions", func(t *testing.T) {

		t.Parallel()

		i1 := &InterfaceType{
			CompositeKind: common.CompositeKindResource,
			Identifier:    "I1",
			Location:      common.StringLocation("b"),
		}

		i2 := &InterfaceType{
			CompositeKind: common.CompositeKindResource,
			Identifier:    "I2",
			Location:      common.StringLocation("b"),
		}

		a := &RestrictedType{
			Type: &CompositeType{
				Kind:       common.CompositeKindResource,
				Identifier: "R1",
				Location:   common.StringLocation("a"),
			},
			Restrictions: []*InterfaceType{i1, i2},
		}

		b := &RestrictedType{
			Type: &CompositeType{
				Kind:       common.CompositeKindResource,
				Identifier: "R2",
				Location:   common.StringLocation("a"),
			},
			Restrictions: []*InterfaceType{i1, i2},
		}

		assert.False(t, a.Equal(b))
	})
}

func TestRestrictedType_GetMember(t *testing.T) {

	t.Parallel()

	t.Run("forbid undeclared members", func(t *testing.T) {

		t.Parallel()

		resourceType := &CompositeType{
			Kind:       common.CompositeKindResource,
			Identifier: "R",
			Location:   common.StringLocation("a"),
			Fields:     []string{},
			Members:    NewStringMemberOrderedMap(),
		}
		ty := &RestrictedType{
			Type:         resourceType,
			Restrictions: []*InterfaceType{},
		}

		fieldName := "s"
		resourceType.Members.Set(fieldName, NewPublicConstantFieldMember(
			ty.Type,
			fieldName,
			IntType,
			"",
		))

		actualMembers := ty.GetMembers()

		require.Contains(t, actualMembers, fieldName)

		var reportedError error
		actualMember := actualMembers[fieldName].Resolve(fieldName, ast.Range{}, func(err error) {
			reportedError = err
		})

		assert.IsType(t, &InvalidRestrictedTypeMemberAccessError{}, reportedError)
		assert.NotNil(t, actualMember)
	})

	t.Run("allow declared members", func(t *testing.T) {

		t.Parallel()

		interfaceType := &InterfaceType{
			CompositeKind: common.CompositeKindResource,
			Identifier:    "I",
			Members:       NewStringMemberOrderedMap(),
		}

		resourceType := &CompositeType{
			Kind:       common.CompositeKindResource,
			Identifier: "R",
			Location:   common.StringLocation("a"),
			Fields:     []string{},
			Members:    NewStringMemberOrderedMap(),
		}
		restrictedType := &RestrictedType{
			Type: resourceType,
			Restrictions: []*InterfaceType{
				interfaceType,
			},
		}

		fieldName := "s"

		resourceType.Members.Set(fieldName, NewPublicConstantFieldMember(
			restrictedType.Type,
			fieldName,
			IntType,
			"",
		))

		interfaceMember := NewPublicConstantFieldMember(
			restrictedType.Type,
			fieldName,
			IntType,
			"",
		)
		interfaceType.Members.Set(fieldName, interfaceMember)

		actualMembers := restrictedType.GetMembers()

		require.Contains(t, actualMembers, fieldName)

		actualMember := actualMembers[fieldName].Resolve(fieldName, ast.Range{}, nil)

		assert.Same(t, interfaceMember, actualMember)
	})
}

func TestBeforeType_Strings(t *testing.T) {

	t.Parallel()

	expected := "(<T: AnyStruct>(_ value: T): T)"

	assert.Equal(t,
		expected,
		beforeType.String(),
	)

	assert.Equal(t,
		expected,
		beforeType.QualifiedString(),
	)
}

func TestQualifiedIdentifierCreation(t *testing.T) {

	t.Run("with containers", func(t *testing.T) {

		a := &CompositeType{
			Kind:       common.CompositeKindStructure,
			Identifier: "A",
			Location:   common.StringLocation("a"),
			Fields:     []string{},
			Members:    NewStringMemberOrderedMap(),
		}

		b := &CompositeType{
			Kind:          common.CompositeKindStructure,
			Identifier:    "B",
			Location:      common.StringLocation("a"),
			Fields:        []string{},
			Members:       NewStringMemberOrderedMap(),
			containerType: a,
		}

		c := &CompositeType{
			Kind:          common.CompositeKindStructure,
			Identifier:    "C",
			Location:      common.StringLocation("a"),
			Fields:        []string{},
			Members:       NewStringMemberOrderedMap(),
			containerType: b,
		}

		identifier := qualifiedIdentifier("foo", c)
		assert.Equal(t, "A.B.C.foo", identifier)
	})

	t.Run("without containers", func(t *testing.T) {
		identifier := qualifiedIdentifier("foo", nil)
		assert.Equal(t, "foo", identifier)
	})

	t.Run("public account container", func(t *testing.T) {
		identifier := qualifiedIdentifier("foo", PublicAccountType)
		assert.Equal(t, "PublicAccount.foo", identifier)
	})
}

func BenchmarkQualifiedIdentifierCreation(b *testing.B) {

	foo := &CompositeType{
		Kind:       common.CompositeKindStructure,
		Identifier: "foo",
		Location:   common.StringLocation("a"),
		Fields:     []string{},
		Members:    NewStringMemberOrderedMap(),
	}

	bar := &CompositeType{
		Kind:          common.CompositeKindStructure,
		Identifier:    "bar",
		Location:      common.StringLocation("a"),
		Fields:        []string{},
		Members:       NewStringMemberOrderedMap(),
		containerType: foo,
	}

	b.Run("One level", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			qualifiedIdentifier("baz", nil)
		}
	})

	b.Run("Three levels", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			qualifiedIdentifier("baz", bar)
		}
	})
}

func TestIdentifierCacheUpdate(t *testing.T) {

	code := `
          pub contract interface Test {

              pub struct interface NestedInterface {
                  pub fun test(): Bool
              }

              pub struct Nested: NestedInterface {}
          }

          pub contract TestImpl {

              pub struct Nested {
                  pub fun test(): Bool {
                      return true
                  }
              }
          }
	`

	program, err := parser2.ParseProgram(code)
	require.NoError(t, err)

	checker, err := NewChecker(
		program,
		common.StringLocation("test"),
	)
	require.NoError(t, err)

	err = checker.Check()
	require.NoError(t, err)

	checker.typeActivations.ForEachVariableDeclaredInAndBelow(
		0,
		func(_ string, value *Variable) {
			typ := value.Type

			var checkIdentifiers func(t *testing.T, typ Type)

			checkNestedTypes := func(nestedTypes *StringTypeOrderedMap) {
				if nestedTypes != nil {
					nestedTypes.Foreach(
						func(_ string, typ Type) {
							checkIdentifiers(t, typ)
						},
					)
				}
			}

			checkIdentifiers = func(t *testing.T, typ Type) {
				switch semaType := typ.(type) {
				case *CompositeType:
					cachedQualifiedID := semaType.QualifiedIdentifier()
					cachedID := semaType.ID()

					// clear cached identifiers for one level
					semaType.cachedIdentifiers = nil

					recalculatedQualifiedID := semaType.QualifiedIdentifier()
					recalculatedID := semaType.ID()

					assert.Equal(t, recalculatedQualifiedID, cachedQualifiedID)
					assert.Equal(t, recalculatedID, cachedID)

					// Recursively check for nested types
					checkNestedTypes(semaType.nestedTypes)

				case *InterfaceType:
					cachedQualifiedID := semaType.QualifiedIdentifier()
					cachedID := semaType.ID()

					// clear cached identifiers for one level
					semaType.cachedIdentifiers = nil

					recalculatedQualifiedID := semaType.QualifiedIdentifier()
					recalculatedID := semaType.ID()

					assert.Equal(t, recalculatedQualifiedID, cachedQualifiedID)
					assert.Equal(t, recalculatedID, cachedID)

					// Recursively check for nested types
					checkNestedTypes(semaType.nestedTypes)
				}
			}

			checkIdentifiers(t, typ)
		})
}

func TestCommonSuperType(t *testing.T) {

	t.Run("duplicate mask", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				err, _ := r.(error)
				require.Error(t, err)
				assert.Equal(t, "duplicate type tag: {32 0}", err.Error())
			}
		}()

		_ = newTypeTagFromLowerMask(32)
	})

	nilType := &OptionalType{NeverType}

	resourceType := &CompositeType{
		Location:   nil,
		Identifier: "Foo",
		Kind:       common.CompositeKindResource,
	}

	fmt.Println(CommonSuperType(UInt8Type, UInt8Type, UInt8Type))
	fmt.Println(CommonSuperType(UInt8Type, UInt16Type, UInt256Type))
	fmt.Println(CommonSuperType(Int8Type, Int16Type))

	fmt.Println(CommonSuperType(nilType, nilType))
	fmt.Println(CommonSuperType(Int8Type, nilType))

	fmt.Println(CommonSuperType(AnyStructType, AnyStructType))
	fmt.Println(CommonSuperType(AnyResourceType, AnyResourceType))
	fmt.Println(CommonSuperType(AnyStructType, AnyResourceType))

	fmt.Println(CommonSuperType(Int8Type, StringType))
	fmt.Println(CommonSuperType(nilType, StringType))

	fmt.Println("---- Arrays ---- ")
	fmt.Println(CommonSuperType(
		&VariableSizedType{Type: StringType},
		&VariableSizedType{Type: StringType},
	))

	fmt.Println(CommonSuperType(
		&VariableSizedType{Type: StringType},
		&ConstantSizedType{Type: StringType, Size: 2},
	))

	fmt.Println(CommonSuperType(
		&VariableSizedType{Type: StringType},
		&VariableSizedType{Type: BoolType},
	))

	fmt.Println(CommonSuperType(
		&VariableSizedType{Type: StringType},
		StringType,
	))

	fmt.Println(CommonSuperType(
		&VariableSizedType{Type: StringType},
		&VariableSizedType{Type: resourceType},
	))

	fmt.Println(CommonSuperType(
		&VariableSizedType{Type: resourceType},
		&VariableSizedType{Type: resourceType},
	))

	fmt.Println(CommonSuperType(
		resourceType,
		&VariableSizedType{Type: resourceType},
	))

	fmt.Println(CommonSuperType(
		&VariableSizedType{Type: &VariableSizedType{Type: resourceType}},
		&VariableSizedType{Type: &VariableSizedType{Type: resourceType}},
	))

	fmt.Println(CommonSuperType(
		&VariableSizedType{Type: &VariableSizedType{Type: resourceType}},
		&VariableSizedType{Type: &VariableSizedType{Type: StringType}},
	))

	fmt.Println("---- Dictionaries ---- ")
	stringStringDictionary := &DictionaryType{
		KeyType:   StringType,
		ValueType: StringType,
	}

	stringBoolDictionary := &DictionaryType{
		KeyType:   StringType,
		ValueType: BoolType,
	}

	stringResourceDictionary := &DictionaryType{
		KeyType:   StringType,
		ValueType: resourceType,
	}

	assertLeastCommonSuperType := func(expectedType Type, types ...Type) {
		assert.Equal(t, expectedType, CommonSuperType(types...))
	}

	assertLeastCommonSuperType(
		stringStringDictionary,
		stringStringDictionary,
		stringStringDictionary,
	)

	assertLeastCommonSuperType(
		AnyStructType,
		stringStringDictionary,
		stringBoolDictionary,
	)

	assertLeastCommonSuperType(
		AnyStructType,
		stringStringDictionary,
		StringType,
	)

	assertLeastCommonSuperType(
		NeverType,
		stringStringDictionary,
		stringResourceDictionary,
	)

	assertLeastCommonSuperType(
		stringResourceDictionary,
		stringResourceDictionary,
		stringResourceDictionary,
	)

	assertLeastCommonSuperType(
		AnyResourceType,
		resourceType,
		stringResourceDictionary,
	)

	nestedResourceDictionary := &DictionaryType{
		KeyType:   StringType,
		ValueType: stringResourceDictionary,
	}

	nestedStringDictionary := &DictionaryType{
		KeyType:   StringType,
		ValueType: stringStringDictionary,
	}

	assertLeastCommonSuperType(
		nestedResourceDictionary,
		nestedResourceDictionary,
		nestedResourceDictionary,
	)

	assertLeastCommonSuperType(
		NeverType,
		nestedStringDictionary,
		nestedResourceDictionary,
	)
}
