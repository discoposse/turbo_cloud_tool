package turbo

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
)

type CostTopoFile struct {
	Tables []Table
}

func NewCostTopoFile(filename string) (*CostTopoFile, error) {
	var retval *CostTopoFile
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return retval, err
	}

	tables := make([]Table, 0)

	err = json.Unmarshal(file, &tables)
	if err != nil {
		return retval, err
	}

	retval = &CostTopoFile{
		Tables: tables,
	}

	return retval, nil
}

func (f CostTopoFile) GetDataCenters() ([]Datacenter, error) {
	retval := make([]Datacenter, 0)
	for _, table := range f.Tables {
		if table.Name == "DataCenters" {
			for _, outer_ary := range table.Items {
				for _, item := range outer_ary {
					// Do the dirty thing of converting this *back* to JSON, then
					// converting it back to the desired object type.
					dc_json, err := json.Marshal(item)
					// TODO: Can't really ignore errors here, but should they be returned?
					if err != nil {
						return retval, err
					}
					var dc Datacenter
					err = json.Unmarshal(dc_json, &dc)
					if err != nil {
						return retval, err
					}
					retval = append(retval, dc)
				}
			}
		}
	}
	return retval, nil
}

func (f CostTopoFile) GetProfiles() ([]Profile, error) {
	retval := make([]Profile, 0)
	for _, table := range f.Tables {
		if table.Name == "Profiles" {
			for _, outer_ary := range table.Items {
				for _, item := range outer_ary {
					// Do the dirty thing of converting this *back* to JSON, then
					// converting it back to the desired object type.
					p_json, err := json.Marshal(item)
					// TODO: Can't really ignore errors here, but should they be returned?
					if err != nil {
						return retval, err
					}
					var p Profile
					err = json.Unmarshal(p_json, &p)
					if err != nil {
						return retval, err
					}
					retval = append(retval, p)
				}
			}
		}
	}
	return retval, nil
}

func (f CostTopoFile) GetOnDemandCosts() ([]OnDemandCost, error) {
	retval := make([]OnDemandCost, 0)
	for _, table := range f.Tables {
		if table.Name == "onDemandCost" {
			for _, item := range table.Items {
				// Do the dirty thing of converting this *back* to JSON, then
				// converting it back to the desired object type.
				o_json, err := json.Marshal(item)
				// TODO: Can't really ignore errors here, but should they be returned?
				if err != nil {
					return retval, err
				}
				var o OnDemandCost
				err = json.Unmarshal(o_json, &o)
				if err != nil {
					return retval, err
				}
				retval = append(retval, o)
			}
		}
	}
	return retval, nil
}

type Table struct {
	Name  string                   `json:"tableName,omitempty"`
	Items []map[string]interface{} `json:"items"`
}

type Datacenter struct {
	Name string `json:"name,omitempty"`
}

type OnDemandCost struct {
	Datacenter string `json:"datacenter,omitempty"`
	Template   string `json:"template,omitempty"`
	Cost       string `json:"cost,omitempty"`
}

func (o OnDemandCost) GetCost() float64 {
	retval := 0.0
	retval, _ = strconv.ParseFloat(o.Cost, 64)
	return retval
}

type Profile struct {
	Name       string          `json:"name,omitempty"`
	Properties ProfileProperty `json:"properties,omitempty"`
}

func (p Profile) GetCpuCapacity() float64 {
	return p.Properties.VMPROFILE_VCPU_SPEED_MHZ * p.Properties.VMPROFILE_VCPU_COUNT
}

func (p Profile) FindClosestMatch(other_profiles []Profile, dest_costs []OnDemandCost, margin int) (Profile, []Profile) {
	float_margin := float64(margin) / 100.00
	candidates := make([]Profile, 0)
	this_cpu := p.GetCpuCapacity()
	this_cpu_min := this_cpu - this_cpu*float_margin
	this_cpu_max := this_cpu + this_cpu*float_margin
	this_mem := p.Properties.VMPROFILE_VMEM_SIZE_KB
	this_mem_min := this_mem - this_mem*float_margin
	this_mem_max := this_mem + this_mem*float_margin
	// Look for an exact match on CPU and Memory
	for _, other_profile := range other_profiles {
		if this_cpu == other_profile.GetCpuCapacity() &&
			this_mem == other_profile.Properties.VMPROFILE_VMEM_SIZE_KB {
			return other_profile, []Profile{other_profile}
		}
	}

	// Look for an exact match on CPU, and a 10% under/over match on Memory
	for _, other_profile := range other_profiles {
		if this_cpu == other_profile.GetCpuCapacity() &&
			this_mem_min < other_profile.Properties.VMPROFILE_VMEM_SIZE_KB &&
			this_mem_max > other_profile.Properties.VMPROFILE_VMEM_SIZE_KB {
			candidates = append(candidates, other_profile)
		}
	}

	if len(candidates) == 0 {
		// Look for a 10% under/over match on Memory and CPU
		for _, other_profile := range other_profiles {
			if this_cpu_min < other_profile.GetCpuCapacity() &&
				this_cpu_max > other_profile.GetCpuCapacity() &&
				this_mem_min < other_profile.Properties.VMPROFILE_VMEM_SIZE_KB &&
				this_mem_max > other_profile.Properties.VMPROFILE_VMEM_SIZE_KB {
				candidates = append(candidates, other_profile)
			}
		}
	}

	var retval Profile
	// Ludicrously high cost for the first check
	last_cost := 10000.0
	for _, candidate := range candidates {
		for _, cost := range dest_costs {
			if cost.Template == candidate.Name && cost.GetCost() < last_cost {
				last_cost = cost.GetCost()
				retval = candidate
			}
		}
	}

	return retval, candidates
}

type ProfileProperty struct {
	VMPROFILE_VCPU_SPEED_MHZ         float64 `json:"VMPROFILE_VCPU_SPEED_MHZ,omitempty"`
	VMPROFILE_VCPU_COUNT             float64 `json:"VMPROFILE_VCPU_COUNT,omitempty"`
	VMPROFILE_IOPS                   float64 `json:"VMPROFILE_IOPS,omitempty"`
	VMPROFILE_PROVIDES_LOCAL_STORAGE bool    `json:"VMPROFILE_PROVIDES_LOCAL_STORAGE,omitempty"`
	VMPROFILE_IS_MATCH               bool    `json:"VMPROFILE_IS_MATCH,omitempty"`
	VMPROFILE_VMEM_SIZE_KB           float64 `json:"VMPROFILE_VMEM_SIZE_KB,omitempty"`
}
