package pingdom

import (
	"fmt"
	"log"
	"strconv"
//	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/Stremio/go-pingdom/pingdom"
)

func resourcePingdomMaintenance() *schema.Resource {
	return &schema.Resource{
		Create: resourcePingdomMaintenanceCreate,
		Read:   resourcePingdomMaintenanceRead,
		Update: resourcePingdomMaintenanceUpdate,
		Delete: resourcePingdomMaintenanceDelete,

		Schema: map[string]*schema.Schema{

			"description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"from": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"to": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"recurrto": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"recurrencetype": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"repeatevery": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
			},

			"uptimecheckids": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"transactioncheckids": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
		},
	}
}

type commonParams struct {
	description string
	from int
	to int
	recurrto int
	recurrencetype string
	repeatevery int
	uptimecheckids string
	transactioncheckids string
}

func checkForMaintenanceResource(d *schema.ResourceData) (commonParams) {
	params := commonParams{}

	if v, ok := d.GetOk("description"); ok {
		params.description = v.(string)
	}
	if v, ok := d.GetOk("from"); ok {
		params.from = v.(int)
	}
	if v, ok := d.GetOk("to"); ok {
		params.to = v.(int)
	}
	if v, ok := d.GetOk("recurrto"); ok {
		params.recurrto = v.(int)
	}
	if v, ok := d.GetOk("recurrencetype"); ok {
		params.recurrencetype = v.(string)
	}
	if v, ok := d.GetOk("repeatevery"); ok {
		params.repeatevery = v.(int)
	}
	if v, ok := d.GetOk("uptimecheckids");ok {
		params.uptimecheckids = v.(string)
	}
	if v, ok := d.GetOk("transactioncheckids"); ok {
		params.transactioncheckids = v.(string)
	}
	return params
}

func reformatDate(input string) int{
	if(input == "now"){
		return 1;
	}

	currentTime := time.Now()
	var currentTimeString string
	currentTimeString = currentTime.Format("02-01-2006 15:04:05")
	var currentYear, currentMonth, currentDay, currentHour, currentMinute, currentSecond int
	currentDay, _ = strconv.Atoi(currentTimeString[0:2])
	currentMonth, _ = strconv.Atoi(currentTimeString[3:5])
	currentYear, _ = strconv.Atoi(currentTimeString[6:10])
	currentHour, _ = strconv.Atoi(currentTimeString[11:13])
	currentMinute, _ = strconv.Atoi(currentTimeString[14:16])
	currentSecond, _ = strconv.Atoi(currentTimeString[17:19])

	var inputYear, inputMonth, inputDay, inputHour, inputMinute, inputSecond int
	inputDay, _ = strconv.Atoi(input[0:2])
	inputMonth, _ = strconv.Atoi(input[3:5])
	inputYear, _ = strconv.Atoi(input[6:10])
	inputHour, _ = strconv.Atoi(input[11:13])
	inputMinute, _ = strconv.Atoi(input[14:16])
	inputSecond, _ = strconv.Atoi(input[17:19])

	var diffYear, diffMonth, diffDay, diffHour, diffMinute, diffSecond int
	diffYear = inputYear - currentYear
	diffMonth = inputMonth - currentMonth
	diffDay = inputDay - currentDay
	diffHour = inputHour - currentHour
	diffMinute = inputMinute - currentMinute
	diffSecond = inputSecond - currentSecond

	location, _ := time.LoadLocation("Europe/Sofia")
	toconv := time.Date(1970, time.January, 1, diffHour, diffMinute, diffSecond, 0, location)
	toconv = toconv.AddDate(diffYear, diffMonth, diffDay)
	return int(toconv.Unix()+7200)

}

func resourcePingdomMaintenanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pingdom.Client)

	params := checkForMaintenanceResource(d)

	log.Printf("[DEBUG] Maintenance create configuration: %#v", d.Get("description"))

	m := pingdom.MaintenanceWindow{
		Description: params.description,
		From:        int64(params.from),
		To:          int64(params.to),
		EffectiveTo: params.recurrto,
		RecurrenceType: params.recurrencetype,
		RepeatEvery: params.repeatevery,
		TmsIDs: params.transactioncheckids,
		UptimeIDs: params.uptimecheckids,

	}

	maintenance, err := client.Maintenances.Create(&m)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(maintenance.ID))

	return nil
}

func resourcePingdomMaintenanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pingdom.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving id for resource: %s", err)
	}
	cl, err := client.Maintenances.List()
	if err != nil {
		return fmt.Errorf("Error retrieving list of checks: %s", err)
	}
	exists := false
	for _, ckid := range cl {
		if ckid.ID == id {
			exists = true
			break
		}
	}
	if !exists {
		d.SetId("")
		return nil
	}
	ck, err := client.Maintenances.Read(id)
	if err != nil {
		return fmt.Errorf("Error retrieving maintenance: %s", err)
	}

	d.Set("description", ck.Description)
	d.Set("from", ck.From)
	d.Set("to", ck.To)
	d.Set("recurrto", ck.EffectiveTo)
	d.Set("recurrencetype", ck.RecurrenceType)
	d.Set("repeatevery", ck.RepeatEvery)

	checks := ck.Checks

	uptimeids := schema.NewSet(
		func(Uptime interface{}) int { return Uptime.(int) },
		[]interface{}{},
	)
	for _, Uptime := range checks.Uptime {
		uptimeids.Add(Uptime)
	}
	d.Set("uptimecheckids", uptimeids)

	transactionids := schema.NewSet(
		func(Transaction interface{}) int { return Transaction.(int) },
		[]interface{}{},
	)
	for _, Transaction := range checks.Tms {
		transactionids.Add(Transaction)
	}
	d.Set("transactioncheckids", transactionids)

	return nil
}

func resourcePingdomMaintenanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pingdom.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving id for resource: %s", err)
	}

	params := checkForMaintenanceResource(d)

	log.Printf("[DEBUG] Maintenance update configuration: %#v", d.Get("description"))

	m := pingdom.MaintenanceWindow{
		Description: params.description,
		From:        int64(params.from),
		To:          int64(params.to),
		EffectiveTo: params.recurrto,
		RecurrenceType: params.recurrencetype,
		RepeatEvery: params.repeatevery,
		TmsIDs: params.transactioncheckids,
		UptimeIDs: params.uptimecheckids,

	}

	_, err = client.Maintenances.Update(id, &m)
	if err != nil {
		return fmt.Errorf("Error updating check: %s", err)
	}

	return nil
}

func resourcePingdomMaintenanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pingdom.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving id for resource: %s", err)
	}

	log.Printf("[DEBUG] Maintenance delete configuration: %#v", d.Get("description"))

	m := pingdom.MaintenanceWindow{
		To:          1,
		EffectiveTo: 1,
	}

	_, err = client.Maintenances.Update(id, &m)
	if err != nil {
		return fmt.Errorf("Error deleting check: %s", err)
	}

	return nil
}
