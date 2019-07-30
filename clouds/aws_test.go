package clouds

import (
	// . "git.turbonomic.com/rgeyer/cloud_pricing_tool"
	// . "git.turbonomic.com/rgeyer/cloud_pricing_tool/clouds/mocks"

	"github.com/aws/aws-sdk-go/service/iam"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWS", func() {
	Describe("AWS User Match", func() {
		Context("Username", func() {
			Context("Matches Exactly", func() {
				It("Matches", func() {
					username := "foo"
					awsUserMatch := AwsUserMatch{
						User: &iam.User{
							UserName: &username,
						},
						Username: "foo",
					}

					Ω(awsUserMatch.UsernameMatch()).To(Equal(true))
				})
			})

			Context("Case insensitive match", func() {
				It("Matches", func() {
					username := "FoO"
					match := AwsUserMatch{
						User: &iam.User{
							UserName: &username,
						},
						Username: "foo",
					}

					Ω(match.UsernameMatch()).To(Equal(true))
				})
			})

			Context("Iam User is Nil", func() {
				It("Does Not Match", func() {
					awsUserMatch := AwsUserMatch{
						Username: "foo",
					}

					Ω(awsUserMatch.UsernameMatch()).To(Equal(false))
				})
			})
		})

		Context("Matched Tags", func() {
			Context("User is nil", func() {
				It("Returns no matches", func() {
					match := AwsUserMatch{}
					Ω(len(match.MatchedTags())).To(Equal(0))
				})
			})

			Context("User has no tags", func() {
				It("Returns no matches", func() {
					match := AwsUserMatch{
						User: &iam.User{},
					}
					Ω(len(match.MatchedTags())).To(Equal(0))
				})
			})

			Context("No query tags defined", func() {
				It("Returns no matches", func() {
					keyvalstr := "both"
					match := AwsUserMatch{
						User: &iam.User{
							Tags: []*iam.Tag{
								&iam.Tag{
									Value: &keyvalstr,
									Key:   &keyvalstr,
								},
							},
						},
					}
					Ω(len(match.MatchedTags())).To(Equal(0))
				})
			})

			Context("String Query Tags Supplied", func() {
				Context("One Exact Match", func() {
					It("Returns exact match", func() {
						match := AwsUserMatch{
							User: &iam.User{
								Tags: ConvertStringTagsPointer([]string{"key:val"}),
							},
							StringTags: []string{"key:val"},
						}

						Ω(len(match.MatchedTags())).To(Equal(1))
						Ω(match.AllTagsMatch()).To(Equal(true))
						Ω(match.AnyTagsMatch()).To(Equal(true))
						Ω(match.AnyMatch()).To(Equal(true))
					})
				})

				Context("One Case Insensitive Match", func() {
					It("Returns exact match", func() {
						match := AwsUserMatch{
							User: &iam.User{
								Tags: ConvertStringTagsPointer([]string{"kEy:VaL"}),
							},
							StringTags: []string{"key:val"},
						}

						Ω(len(match.MatchedTags())).To(Equal(1))
						Ω(match.AllTagsMatch()).To(Equal(true))
						Ω(match.AnyTagsMatch()).To(Equal(true))
					})
				})
			})

			PContext("Matches One Tag", func() {
				It("Returns match, indicating matched tag", func() {
					Fail("Not implemented")
				})
			})

			PContext("Matches Multiple Tags", func() {
				It("Returns match, indicating all matched tags", func() {
					Fail("Not implemented")
				})
			})

			PContext("Matches All Tags", func() {
				It("Returns match, with boolean for exact match", func() {
					Fail("Not implemented")
				})
			})
		})
	})
})
