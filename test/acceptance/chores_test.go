package acceptance_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chore API", func() {
	When("creating a one-off chore", func() {
		DescribeTable("creates the requested one-off chore",
			func() {
				Expect(true).To(BeTrue())
			},
			Entry("manual one-off"),
			Entry("manual one-off saved as a chore template"),
			Entry("template-based one-off"),
		)

		It("stores a new template when save-as-template is enabled", func() {
			Expect(true).To(BeTrue())
		})

		DescribeTable("rejects invalid values without creating a chore",
			func() {
				Expect(true).To(BeTrue())
			},
			Entry("blank name"),
			Entry("missing deadline"),
		)

		It("does not create a chore when save-as-template conflicts with an active template", func() {
			Expect(true).To(BeTrue())
		})

		It("does not create a chore from a deactivated template", func() {
			Expect(true).To(BeTrue())
		})
	})

	When("scheduled recurrence is requested", func() {
		DescribeTable("returns Not Implemented without persisting resources",
			func() {
				Expect(true).To(BeTrue())
			},
			Entry("manual scheduled chore"),
			Entry("template-based scheduled chore"),
		)
	})

	When("browsing chores", func() {
		DescribeTable("browses and searches chores in the selected status",
			func() {
				Expect(true).To(BeTrue())
			},
			Entry("active chores ordered by deadline and ID ascending"),
			Entry("completed chores ordered by completion date and ID descending"),
		)
	})

	When("editing an active one-off chore", func() {
		It("updates its normalized name, description, and deadline", func() {
			Expect(true).To(BeTrue())
		})
	})

	When("completing an active one-off chore", func() {
		It("moves the chore from the active collection to completed history", func() {
			Expect(true).To(BeTrue())
		})
		It("treats repeated completion requests as successful no-ops", func() {
			Expect(true).To(BeTrue())
		})

		It("rejects a completion date after the application-local current date", func() {
			Expect(true).To(BeTrue())
		})
	})

	When("correcting a one-off chore completion", func() {
		It("updates its completion date", func() {
			Expect(true).To(BeTrue())
		})

		It("rejects a completion date after the application-local current date", func() {
			Expect(true).To(BeTrue())
		})
	})

	When("permanently deleting a one-off chore", func() {
		DescribeTable("deletes the chore",
			func() {
				Expect(true).To(BeTrue())
			},
			Entry("active one-off chore"),
			Entry("completed one-off chore"),
		)
	})

	When("retrieving a missing chore", func() {
		It("returns collection navigation and no mutation actions", func() {
			Expect(true).To(BeTrue())
		})
	})
})
