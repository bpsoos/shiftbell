package demodata_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bpsoos/shiftbell/internal/demodata"
	models "github.com/bpsoos/shiftbell/internal/models/demo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDemoData(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Demo Data Loader Suite")
}

var _ = Describe("Load", func() {
	var path string

	BeforeEach(func() {
		path = filepath.Join(GinkgoT().TempDir(), "data.json")
	})

	It(
		"loads custom data and distinguishes omitted completion from completion today",
		func() {
			Expect(os.WriteFile(path, []byte(`{
			"one_off_chores": [
				{"name":"Buy milk","description":"Two bottles","deadline_days":2,"save_as_chore_template":true},
				{"name":"Call back","deadline_days":-1,"completed_days":0}
			],
			"scheduled_chores": [
				{"name":"Sweep","description":"Kitchen","deadline_days":-3,"schedule_name":"Weekly sweep","interval_days":7}
			]
		}`), 0o600)).To(Succeed())
			data, err := demodata.Load(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal(&models.Data{
				OneOffChores: []models.OneOffChore{
					{
						Name:                "Buy milk",
						Description:         "Two bottles",
						DeadlineDays:        2,
						SaveAsChoreTemplate: true,
					},
					{Name: "Call back", DeadlineDays: -1, CompletedDays: new(0)},
				},
				ScheduledChores: []models.ScheduledChore{
					{
						Name:         "Sweep",
						Description:  "Kitchen",
						DeadlineDays: -3,
						ScheduleName: "Weekly sweep",
						IntervalDays: 7,
					},
				},
			}))
		},
	)

	It("accepts null and omitted descriptions for both kinds of chore", func() {
		Expect(os.WriteFile(path, []byte(`{
			"one_off_chores": [
				{"name":"Null description","description":null},
				{"name":"Omitted description"}
			],
			"scheduled_chores": [
				{"name":"Null description","description":null,"schedule_name":"First","interval_days":7},
				{"name":"Omitted description","schedule_name":"Second","interval_days":14}
			]
		}`), 0o600)).To(Succeed())
		data, err := demodata.Load(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(data.OneOffChores).To(HaveLen(2))
		Expect(data.ScheduledChores).To(HaveLen(2))
		Expect(data.OneOffChores[0].Description).To(BeEmpty())
		Expect(data.OneOffChores[1].Description).To(BeEmpty())
		Expect(data.ScheduledChores[0].Description).To(BeEmpty())
		Expect(data.ScheduledChores[1].Description).To(BeEmpty())
	})

	It("requires a path", func() {
		_, err := demodata.Load("")
		Expect(err).To(MatchError("DEMO_DATA_FILEPATH is required"))
	})

	It("returns file read errors", func() {
		_, err := demodata.Load(path)
		Expect(err).To(MatchError(os.ErrNotExist))
	})

	It("allows an intentionally empty dataset", func() {
		Expect(os.WriteFile(path, []byte(`{}`), 0o600)).To(Succeed())
		data, err := demodata.Load(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(&models.Data{}))
	})

	DescribeTable(
		"rejects invalid input before returning any data",
		func(contents string) {
			Expect(os.WriteFile(path, []byte(contents), 0o600)).To(Succeed())
			data, err := demodata.Load(path)
			Expect(err).To(HaveOccurred())
			Expect(data).To(BeNil())
		},
		Entry("empty file", ""),
		Entry("malformed JSON", "{"),
		Entry("null", "null"),
		Entry("array root", "[]"),
		Entry("unknown field", `{"one_off_chore":[]}`),
		Entry(
			"unknown nested field",
			`{"one_off_chores":[{"name":"Chore","deadlin_days":1}]}`,
		),
		Entry("trailing object", `{} {}`),
		Entry("trailing garbage", `{} garbage`),
		Entry(
			"wrong offset type",
			`{"one_off_chores":[{"name":"Chore","deadline_days":"today"}]}`,
		),
		Entry("missing name", `{"one_off_chores":[{}]}`),
		Entry("blank name", `{"one_off_chores":[{"name":" "}]}`),
		Entry(
			"future completion",
			`{"one_off_chores":[{"name":"Chore","completed_days":1}]}`,
		),
		Entry(
			"missing scheduled chore name",
			`{"scheduled_chores":[{"schedule_name":"Weekly","interval_days":7}]}`,
		),
		Entry(
			"missing schedule name",
			`{"scheduled_chores":[{"name":"Chore","interval_days":7}]}`,
		),
		Entry(
			"missing interval",
			`{"scheduled_chores":[{"name":"Chore","schedule_name":"Weekly"}]}`,
		),
		Entry(
			"negative interval",
			`{"scheduled_chores":[{"name":"Chore","schedule_name":"Weekly","interval_days":-1}]}`,
		),
	)
})
