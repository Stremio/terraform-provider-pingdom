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
	from int64
	to int64
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
		params.from = reformatDate(v.(string))
	}
	if v, ok := d.GetOk("to"); ok {
		params.to = reformatDate(v.(string))
	}
	if v, ok := d.GetOk("recurrto"); ok {
		params.recurrto = int(reformatDate(v.(string)))
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

func reformatDate(input string) int64 {
	if input == "now" {
		return time.Now().Unix()
	}

	rtime, error := time.Parse(time.RFC3339, input)

	if error!= nil {
		fmt.Errorf("Invalid time spec %s:  %#v", input, error)
		return -1
	}

	log.Printf("[DEBUG] Maintenance reformatDate %s configuration: %#v", input, rtime)

	return rtime.Unix()
}

func resourcePingdomMaintenanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pingdom.Client)

	params := checkForMaintenanceResource(d)

	log.Printf("[DEBUG] Maintenance create configuration: %#v", d.Get("description"))

	m := pingdom.MaintenanceWindow{
		Description: params.description,
		From:        params.from,
		To:          params.to,
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
		From:        params.from,
		To:          params.to,
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
