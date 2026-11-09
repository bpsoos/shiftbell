package accounts_test

import (
	"errors"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/accounts"
	"github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/bpsoos/shiftbell/internal/service/accounts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {
	var persister *MockPersister
	var validator *MockUsernameValidator
	var passwordValidator *MockPasswordValidator
	var hasher *MockPasswordHasher
	var service *accounts.Service

	BeforeEach(func() {
		persister = &MockPersister{}
		validator = &MockUsernameValidator{}
		passwordValidator = &MockPasswordValidator{}
		hasher = &MockPasswordHasher{}
		DeferCleanup(func() {
			persister.AssertExpectations(GinkgoT())
			validator.AssertExpectations(GinkgoT())
			passwordValidator.AssertExpectations(GinkgoT())
			hasher.AssertExpectations(GinkgoT())
		})
		service = accounts.NewService(&accounts.Deps{
			Persister:         persister,
			UsernameValidator: validator,
			PasswordValidator: passwordValidator,
			PasswordHasher:    hasher,
		})
	})

	It(
		"stores the canonical username and password hash without changing the input",
		func(ctx SpecContext) {
			params := &models.CreateParams{
				Username: "A\u0301lex",
				Password: " a long pa\u0301ssphrase\t ",
			}
			original := *params
			account := &models.Account{Id: 7, Username: "Álex"}
			validator.On("Validate", params.Username).Return("Álex", nil).Once()
			passwordValidator.On("Validate", params.Password).Return(nil).Once()
			hasher.On("Hash", params.Password).Return("encoded-password-hash", nil).Once()
			persister.On("Create", ctx, &models.CreateValues{
				Username:     "Álex",
				PasswordHash: "encoded-password-hash",
			}).Return(account, nil).Once()

			result, err := service.Create(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeIdenticalTo(account))
			Expect(*params).To(Equal(original))
		},
	)

	It("rejects missing input without calling dependencies", func(ctx SpecContext) {
		result, err := service.Create(ctx, nil)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(validation.ErrInvalidUsername))
		Expect(err).To(MatchError(validation.ErrRequired))
	})

	DescribeTable(
		"preserves password validation errors without hashing or persisting",
		func(ctx SpecContext, password string, reason error) {
			validator.On("Validate", "Owner").Return("Owner", nil).Once()
			passwordValidator.On("Validate", password).Return(reason).Once()

			result, err := service.Create(ctx, &models.CreateParams{
				Username: "Owner",
				Password: password,
			})

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(validation.ErrInvalidPassword))
			Expect(err).To(MatchError(reason))
			Expect(err.Error()).To(Equal("invalid password: " + reason.Error()))
		},
		Entry("required", "", validation.ErrRequired),
		Entry("too short", "short", validation.ErrTooShort),
		Entry("invalid encoding", string([]byte{0xff}), validation.ErrInvalidUTF8),
		Entry("descriptive details", "short", fmt.Errorf(
			"%w: password must contain at least 15 characters; received 5",
			validation.ErrTooShort,
		)),
	)

	DescribeTable(
		"preserves username validation errors",
		func(ctx SpecContext, reason error) {
			validator.On("Validate", "invalid").Return("", reason).Once()

			result, err := service.Create(ctx, &models.CreateParams{
				Username: "invalid",
				Password: "a long passphrase",
			})

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(validation.ErrInvalidUsername))
			Expect(err).To(MatchError(reason))
			Expect(err.Error()).To(Equal("invalid username: " + reason.Error()))
		},
		Entry("required", validation.ErrRequired),
		Entry("too long", validation.ErrTooLong),
		Entry("invalid encoding", validation.ErrInvalidUTF8),
		Entry("disallowed character", validation.ErrDisallowedCharacter),
		Entry("descriptive details", fmt.Errorf(
			"%w: whitespace U+0020 at character 4 is not allowed",
			validation.ErrDisallowedCharacter,
		)),
	)

	It("does not persist credentials when hashing fails", func(ctx SpecContext) {
		failure := errors.New("hashing failed")
		validator.On("Validate", "Owner").Return("Owner", nil).Once()
		passwordValidator.On("Validate", "a long passphrase").Return(nil).Once()
		hasher.On("Hash", "a long passphrase").Return("", failure).Once()

		result, err := service.Create(ctx, &models.CreateParams{
			Username: "Owner",
			Password: "a long passphrase",
		})

		Expect(result).To(BeNil())
		Expect(errors.Is(err, failure)).To(BeTrue())
	})

	DescribeTable("preserves persistence errors", func(ctx SpecContext, failure error) {
		validator.On("Validate", "Owner").Return("Owner", nil).Once()
		passwordValidator.On("Validate", "a long passphrase").Return(nil).Once()
		hasher.On("Hash", "a long passphrase").Return("encoded-password-hash", nil).Once()
		persister.On("Create", ctx, &models.CreateValues{
			Username:     "Owner",
			PasswordHash: "encoded-password-hash",
		}).Return(nil, failure).Once()

		result, err := service.Create(ctx, &models.CreateParams{
			Username: "Owner",
			Password: "a long passphrase",
		})

		Expect(result).To(BeNil())
		Expect(errors.Is(err, failure)).To(BeTrue())
	},
		Entry("account already exists", models.ErrAlreadyExists),
		Entry("storage failure", errors.New("storage failed")),
	)
})
