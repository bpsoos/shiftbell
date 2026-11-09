package passwordhash_test

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/bpsoos/shiftbell/internal/passwordhash"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/argon2"
)

var _ = Describe("Hasher", func() {
	var hasher *passwordhash.Hasher

	BeforeEach(func() {
		hasher = passwordhash.NewHasher()
	})

	It("encodes everything needed to verify the exact password using Argon2id", func() {
		password := " pa\u0301ssword\t "
		encoded, err := hasher.Hash(password)
		Expect(err).NotTo(HaveOccurred())
		parts := strings.Split(encoded, "$")
		Expect(parts).To(HaveLen(6))
		Expect(parts[:3]).To(Equal([]string{"", "argon2id", "v=19"}))

		var memory, iterations uint32
		var parallelism uint8
		count, err := fmt.Sscanf(
			parts[3],
			"m=%d,t=%d,p=%d",
			&memory,
			&iterations,
			&parallelism,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(count).To(Equal(3))
		Expect(memory).To(Equal(uint32(64 * 1024)))
		Expect(iterations).To(Equal(uint32(3)))
		Expect(parallelism).To(Equal(uint8(1)))
		salt, err := base64.RawStdEncoding.DecodeString(parts[4])
		Expect(err).NotTo(HaveOccurred())
		Expect(salt).To(HaveLen(16))
		key, err := base64.RawStdEncoding.DecodeString(parts[5])
		Expect(err).NotTo(HaveOccurred())
		Expect(key).To(HaveLen(32))

		expected := argon2.IDKey(
			[]byte(password),
			salt,
			iterations,
			memory,
			parallelism,
			uint32(len(key)),
		)
		Expect(key).To(Equal(expected))
		wrong := argon2.IDKey(
			[]byte("wrong password"),
			salt,
			iterations,
			memory,
			parallelism,
			uint32(len(key)),
		)
		Expect(key).NotTo(Equal(wrong))
	})

	It("generates a fresh salt for each hash of the same password", func() {
		first, err := hasher.Hash("same password")
		Expect(err).NotTo(HaveOccurred())
		second, err := hasher.Hash("same password")
		Expect(err).NotTo(HaveOccurred())
		firstParts := strings.Split(first, "$")
		secondParts := strings.Split(second, "$")
		Expect(firstParts).To(HaveLen(6))
		Expect(secondParts).To(HaveLen(6))
		Expect(firstParts[4]).NotTo(Equal(secondParts[4]))
		Expect(firstParts[5]).NotTo(Equal(secondParts[5]))
	})
})
