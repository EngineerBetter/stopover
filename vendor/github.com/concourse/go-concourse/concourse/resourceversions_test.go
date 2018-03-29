package concourse_test

import (
	"fmt"
	"net/http"

	"github.com/concourse/atc"
	"github.com/concourse/go-concourse/concourse"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("ATC Handler Resource Versions", func() {
	Describe("ResourceVersions", func() {
		expectedURL := fmt.Sprint("/api/v1/teams/some-team/pipelines/mypipeline/resources/myresource/versions")

		var expectedVersions []atc.VersionedResource

		var page concourse.Page

		var versions []atc.VersionedResource
		var pagination concourse.Pagination
		var found bool
		var clientErr error

		BeforeEach(func() {
			page = concourse.Page{}

			expectedVersions = []atc.VersionedResource{
				{
					Version: atc.Version{"version": "v1"},
				},
				{
					Version: atc.Version{"version": "v2"},
				},
			}
		})

		JustBeforeEach(func() {
			versions, pagination, found, clientErr = team.ResourceVersions("mypipeline", "myresource", page)
		})

		Context("when since, until, and limit are 0", func() {
			BeforeEach(func() {
				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL),
						ghttp.RespondWithJSONEncoded(http.StatusOK, expectedVersions),
					),
				)
			})

			It("calls to get all versions", func() {
				Expect(clientErr).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(versions).To(Equal(expectedVersions))
			})
		})

		Context("when since is specified", func() {
			BeforeEach(func() {
				page = concourse.Page{Since: 24}

				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL, "since=24"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, expectedVersions),
					),
				)
			})

			It("calls to get all versions since that id", func() {
				Expect(clientErr).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(versions).To(Equal(expectedVersions))
			})
		})

		Context("when since and limit is specified", func() {
			BeforeEach(func() {
				page = concourse.Page{Since: 24, Limit: 5}

				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL, "since=24&limit=5"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, expectedVersions),
					),
				)
			})

			It("appends limit to the url", func() {
				Expect(clientErr).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(versions).To(Equal(expectedVersions))
			})
		})

		Context("when until is specified", func() {
			BeforeEach(func() {
				page = concourse.Page{Until: 26}

				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL, "until=26"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, expectedVersions),
					),
				)
			})

			It("calls to get all versions until that id", func() {
				Expect(clientErr).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(versions).To(Equal(expectedVersions))
			})
		})

		Context("when until and limit is specified", func() {
			BeforeEach(func() {
				page = concourse.Page{Until: 26, Limit: 15}

				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL, "until=26&limit=15"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, expectedVersions),
					),
				)
			})

			It("appends limit to the url", func() {
				Expect(clientErr).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(versions).To(Equal(expectedVersions))
			})
		})

		Context("when since and until are both specified", func() {
			BeforeEach(func() {
				page = concourse.Page{Since: 24, Until: 26}

				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL, "until=26"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, expectedVersions),
					),
				)
			})

			It("only sends the until", func() {
				Expect(clientErr).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(versions).To(Equal(expectedVersions))
			})
		})

		Context("when the server returns an error", func() {
			BeforeEach(func() {
				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL),
						ghttp.RespondWith(http.StatusInternalServerError, ""),
					),
				)
			})

			It("returns false and an error", func() {
				Expect(clientErr).To(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when the server returns not found", func() {
			BeforeEach(func() {
				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL),
						ghttp.RespondWith(http.StatusNotFound, ""),
					),
				)
			})

			It("returns false and no error", func() {
				Expect(clientErr).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Describe("pagination data", func() {
			Context("with a link header", func() {
				BeforeEach(func() {
					atcServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", expectedURL),
							ghttp.RespondWithJSONEncoded(http.StatusOK, expectedVersions, http.Header{
								"Link": []string{
									`<http://some-url.com/api/v1/teams/some-team/pipelines/mypipeline/resources/myresource/versions?since=452&limit=123>; rel="previous"`,
									`<http://some-url.com/api/v1/teams/some-team/pipelines/mypipeline/resources/myresource/versions?until=254&limit=456>; rel="next"`,
								},
							}),
						),
					)
				})

				It("returns the pagination data from the header", func() {
					Expect(clientErr).ToNot(HaveOccurred())
					Expect(pagination.Previous).To(Equal(&concourse.Page{Since: 452, Limit: 123}))
					Expect(pagination.Next).To(Equal(&concourse.Page{Until: 254, Limit: 456}))
				})
			})
		})

		Context("without a link header", func() {
			BeforeEach(func() {
				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", expectedURL),
						ghttp.RespondWithJSONEncoded(http.StatusOK, expectedVersions, http.Header{}),
					),
				)
			})

			It("returns pagination data with nil pages", func() {
				Expect(clientErr).ToNot(HaveOccurred())

				Expect(pagination.Previous).To(BeNil())
				Expect(pagination.Next).To(BeNil())
			})
		})
	})
})
