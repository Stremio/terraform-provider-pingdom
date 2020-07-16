package pingdom

import (
	"fmt"
	"strconv"
	//	"strings"
	"testing"

	//	"github.com/hashicorp/terraform/helper/schema"
	"github.com/Stremio/go-pingdom/pingdom"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestResourcePingdomMaintenance_basic(t *testing.T) {
	var maintenance pingdom.MaintenanceWindow

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { TestProviderConfigure(t) },
		Providers:    testAccProviders,
		CheckDestroy: testResourcePingdomMaintenanceCheckDestroy(&maintenance),
		Steps: []resource.TestStep{
			{
				Config: testMaintenanceTestConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testResourcePingdomMaintenanceCheckExists("pingdom_maintenance.test_downtime", &maintenance),
					testResourcePingdomMaintenanceCheckAttributes("pingdom_maintenance.test_downtime", &maintenance),
				),
			},
		},
	})
}

const testMaintenanceTestConfig_basic = `
resource "pingdom_maintenance" "test_downtime" {
	description = "Simple pingdom test maintenance"
	from = "2020-07-16:12:00:00"
	to = "2020-07-16:13:00:00"
}
`

func testResourcePingdomMaintenanceCheckDestroy(maintenance *pingdom.MaintenanceWindow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*pingdom.Client)
		maintenances, err := client.Maintenances.List()

		id := -1
		for _, tv := range maintenances {
			if tv.Description == maintenance.Description {
				id = tv.ID
				break
			}
		}

		if id == -1 {
			return fmt.Errorf("no such maintenance to delete found")
		}

		_, err = client.Maintenances.Delete(id)
		if err == nil {
			return fmt.Errorf("maintenance still exists")
		}

		return nil
	}
}

func testResourcePingdomMaintenanceCheckExists(rn string, maintenance *pingdom.MaintenanceWindow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("TestID not set")
		}

		client := testAccProvider.Meta().(*pingdom.Client)
		mId, parseErr := strconv.Atoi(rs.Primary.ID)
		if parseErr != nil {
			return fmt.Errorf("error in pingdom maintenance window CheckExists: %s", parseErr)
		}

		gotMaintenance, err := client.Maintenances.Read(mId)
		if err != nil {
			return fmt.Errorf("error getting maintenance window: %s", err)
		}

		*maintenance = pingdom.MaintenanceWindow{
			Description:    gotMaintenance.Description,
			From:           int64(gotMaintenance.From),
			To:             int64(gotMaintenance.To),
			EffectiveTo:    int(gotMaintenance.EffectiveTo),
			RecurrenceType: gotMaintenance.RecurrenceType,
			RepeatEvery:    gotMaintenance.RepeatEvery,
			// 			TmsIDs: gotMaintenance.Checks,
			// 			UptimeIDs: gotMaintenance.Checks,
		}

		return nil
	}
}

func testResourcePingdomMaintenanceCheckAttributes(rn string, maintenance *pingdom.MaintenanceWindow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attrs := s.RootModule().Resources[rn].Primary.Attributes

		check := func(key, stateValue, testValue string) error {
			if testValue != stateValue {
				return fmt.Errorf("different values for %s in state (%s) and in pingdom (%s)",
					key, stateValue, testValue)
			}
			return nil
		}

		for key, value := range attrs {
			var err error

			switch key {
			case "Description":
				err = check(key, value, maintenance.Description)
			case "From":
				err = check(key, value, strconv.FormatInt(maintenance.From, 10))
			case "To":
				err = check(key, value, strconv.FormatInt(maintenance.To, 10))
			case "EffectiveTo":
				err = check(key, value, strconv.FormatInt(int64(maintenance.EffectiveTo), 10))
			case "RecurrenceType":
				err = check(key, value, maintenance.RecurrenceType)
			case "RepeatEvery":
				err = check(key, value, strconv.Itoa(maintenance.RepeatEvery))
			case "TmsIDs":
				for _, tv := range maintenance.TmsIDs {
					err = check(key, value, strconv.FormatInt(int64(tv), 10))
					if err != nil {
						return err
					}
				}
			case "UptimeIDs":
				for _, tv := range maintenance.UptimeIDs {
					err = check(key, value, strconv.FormatInt(int64(tv), 10))
					if err != nil {
						return err
					}
				}
			}
			if err != nil {
				return err
			}
		}
		return nil
	}
}
