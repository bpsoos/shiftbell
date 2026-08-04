package schedules_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSchedulesService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Schedules Service Suite")
}
