package turbo_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "git.turbonomic.com/rgeyer/cloud_pricing_tool/turbo"
)

var _ = Describe("Turbo", func() {
	Context("Cost File", func() {
		Context("Loading file", func() {
			Context("File Exists", func() {
				It("Unmarshals Tables", func() {
					costfile, err := NewCostTopoFile("/Volumes/RJGShared/Turbo/Zebra/topo/srv/tomcat/data/repos/awsv2.cost.topology")
					Ω(err).NotTo(HaveOccurred())
					for _, table := range costfile.Tables {
						fmt.Println(table.Name)
						if table.Name == "DataCenters" {
							fmt.Println(table.Items)
						}
					}
				})
			})
		})

		It("Unmarshals Datacenters", func() {
			costfile, err := NewCostTopoFile("/Volumes/RJGShared/Turbo/Zebra/topo/srv/tomcat/data/repos/awsv2.cost.topology")
			Ω(err).NotTo(HaveOccurred())
			dcs, err := costfile.GetDataCenters()
			Ω(err).NotTo(HaveOccurred())
			Ω(len(dcs)).ShouldNot(Equal(0))
			for _, dc := range dcs {
				fmt.Println(dc.Name)
			}
		})

		It("Unmarshals Profiles", func() {
			costfile, err := NewCostTopoFile("/Volumes/RJGShared/Turbo/Zebra/topo/srv/tomcat/data/repos/awsv2.cost.topology")
			Ω(err).NotTo(HaveOccurred())
			profiles, err := costfile.GetProfiles()
			Ω(err).NotTo(HaveOccurred())
			Ω(len(profiles)).ShouldNot(Equal(0))
			for _, p := range profiles {
				fmt.Println(p.Name)
			}
		})
	})
})
