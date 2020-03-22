package checker

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dapperlabs/cadence/runtime/common"
	"github.com/dapperlabs/cadence/runtime/sema"
	"github.com/dapperlabs/cadence/runtime/stdlib"
	. "github.com/dapperlabs/cadence/runtime/tests/utils"
)

var accountValueDeclaration = stdlib.StandardLibraryValue{
	Name:       "account",
	Type:       &sema.AuthAccountType{},
	Kind:       common.DeclarationKindConstant,
	IsConstant: true,
}

func ParseAndCheckAccount(t *testing.T, code string) (*sema.Checker, error) {
	return ParseAndCheckWithOptions(t,
		code,
		ParseAndCheckOptions{
			Options: []sema.Option{
				sema.WithPredeclaredValues(map[string]sema.ValueDeclaration{
					"account": accountValueDeclaration,
				}),
			},
		},
	)
}

func TestCheckAccount(t *testing.T) {

	t.Run("storage is assignable", func(t *testing.T) {
		_, err := ParseAndCheckAccount(t,
			`
              resource R {}

              fun test(): @R? {
                  let r <- account.storage[R] <- create R()
                  return <-r
              }
            `,
		)

		require.NoError(t, err)
	})

	t.Run("published is assignable", func(t *testing.T) {
		_, err := ParseAndCheckAccount(t,
			`
              resource R {}

              fun test() {
                  account.published[&R] = &account.storage[R] as &R
              }
            `,
		)

		require.NoError(t, err)
	})

	t.Run("saving resources: implicit type argument", func(t *testing.T) {
		_, err := ParseAndCheckAccount(t,
			`
              resource R {}

              fun test() {
                  let r <- create R()
                  account.save(<-r, to: /storage/r)
              }
            `,
		)

		require.NoError(t, err)
	})

	t.Run("saving resources: explicit type argument", func(t *testing.T) {
		_, err := ParseAndCheckAccount(t,
			`
              resource R {}

              fun test() {
                  let r <- create R()
                  account.save<R>(<-r, to: /storage/r)
              }
            `,
		)

		require.NoError(t, err)
	})

	t.Run("saving resources: explicit type argument, incorect", func(t *testing.T) {
		_, err := ParseAndCheckAccount(t,
			`
              resource R {}

              resource T {}

              fun test() {
                  let r <- create R()
                  account.save<T>(<-r, to: /storage/r)
              }
            `,
		)

		errs := ExpectCheckerErrors(t, err, 2)

		require.IsType(t, &sema.TypeParameterTypeMismatchError{}, errs[0])
		require.IsType(t, &sema.TypeMismatchError{}, errs[1])
	})
}
