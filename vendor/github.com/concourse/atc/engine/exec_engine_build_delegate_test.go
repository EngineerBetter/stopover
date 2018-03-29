package engine_test

import (
	"errors"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/concourse/atc"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/db/dbfakes"
	. "github.com/concourse/atc/engine"
	"github.com/concourse/atc/event"
	"github.com/concourse/atc/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildDelegate", func() {
	var (
		factory BuildDelegateFactory

		fakeBuild *dbfakes.FakeBuild

		delegate BuildDelegate

		logger *lagertest.TestLogger

		originID event.OriginID
	)

	BeforeEach(func() {
		factory = NewBuildDelegateFactory()

		fakeBuild = new(dbfakes.FakeBuild)
		delegate = factory.Delegate(fakeBuild)

		logger = lagertest.NewTestLogger("test")

		originID = event.OriginID("some-origin-id")
	})

	Describe("Finish", func() {
		Context("when build was aborted", func() {
			BeforeEach(func() {
				delegate.Finish(logger, nil, exec.Success(false), true)
			})

			It("updates build status to aborted", func() {
				finishedStatus := fakeBuild.FinishArgsForCall(0)
				Expect(finishedStatus).To(Equal(db.BuildStatusAborted))
			})
		})

		Context("when build had error", func() {
			BeforeEach(func() {
				delegate.Finish(logger, errors.New("disaster"), exec.Success(false), false)
			})

			It("updates build status to errorred", func() {
				finishedStatus := fakeBuild.FinishArgsForCall(0)
				Expect(finishedStatus).To(Equal(db.BuildStatusErrored))
			})
		})

		Context("when build succeeded", func() {
			BeforeEach(func() {
				dbBuildDelegate1 := delegate.DBActionsBuildEventsDelegate(atc.PlanID(1))
				dbBuildDelegate1.ActionCompleted(logger, &exec.GetAction{
					Resource: "some-resource-1",
					Type:     "some-resource-type-1",
					VersionSource: &exec.StaticVersionSource{
						Version: atc.Version{"version": "some-version-1"},
					},
				})
				dbBuildDelegate2 := delegate.DBActionsBuildEventsDelegate(atc.PlanID(2))
				dbBuildDelegate2.ActionCompleted(logger, &exec.GetAction{
					Resource: "some-resource-2",
					Type:     "some-resource-type-2",
					VersionSource: &exec.StaticVersionSource{
						Version: atc.Version{"version": "some-version-2"},
					},
				})
				delegate.Finish(logger, nil, exec.Success(true), false)
			})

			It("updates build status to succeeded", func() {
				finishedStatus := fakeBuild.FinishArgsForCall(0)
				Expect(finishedStatus).To(Equal(db.BuildStatusSucceeded))
			})

			It("saves all implicit outputs generated by delegates", func() {
				versionedResource1, explicit := fakeBuild.SaveOutputArgsForCall(0)
				Expect(explicit).To(BeFalse())

				versionedResource2, explicit := fakeBuild.SaveOutputArgsForCall(1)
				Expect(explicit).To(BeFalse())

				Expect([]string{versionedResource1.Resource, versionedResource2.Resource}).To(
					ConsistOf([]string{"some-resource-1", "some-resource-2"}),
				)
				Expect([]string{versionedResource1.Type, versionedResource2.Type}).To(
					ConsistOf([]string{"some-resource-type-1", "some-resource-type-2"}),
				)
			})
		})
	})
})
