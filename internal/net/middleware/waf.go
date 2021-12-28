package middleware

import (
	"fmt"
	"strings"
	"sync"

	"github.com/blackcrw/wprecon/internal/database"
	"github.com/blackcrw/wprecon/internal/models"
	"github.com/blackcrw/wprecon/internal/views"
)

type _waf struct {
	net   *models.ResponseModel
	model *models.MiddlewareFirewallModel
	wg    sync.WaitGroup
}

func ActiveWebApplicationFirewall(model_response *models.ResponseModel) {
	var waf = _waf{ net: model_response }

	var model_waf_response = waf.all()

	if model_waf_response.Confidence > 0 && !database.Memory.GetBool("Middleware Firewall Passing") {
		views.MiddlewareWAFActive(model_waf_response)

		go database.Memory.SetBool("Middleware Firewall Passing", true)
	}
}

func (this *_waf) format_slice_to_string(slice []string) string {	
	if len(slice) == 1 {
		return slice[0]
	} else if len(slice) > 1 {
		return strings.Title(fmt.Sprintf("%s And %s", slice[0], slice[1]))
	}

	return ""
}

func (this *_waf) cloudflare() {
	var (
		name = "Cloudflare"
		found_by_slice []string = []string{}
		confidence int
	)

	if strings.Contains(this.net.Raw, "Cloudflare Ray ID:") || strings.Contains(this.net.Raw, "Attention Required!") || strings.Contains(this.net.Raw, "DDoS protection by Cloudflare") {
		found_by_slice = append(found_by_slice, "Text snippet")
		confidence += 80
	}
	
	if this.net.Response.Header.Get("Server") == "cloudflare" {
		found_by_slice = append(found_by_slice, "Header field value")
		confidence += 20
	}

	defer this.wg.Done()

	this.model = &models.MiddlewareFirewallModel{Name: name, FoundBy: this.format_slice_to_string(found_by_slice), Confidence: confidence}
}

func (this *_waf) cerber() {
	var (
		name = "Wordpress Cerber"
		confidence int
		solve_slice []string = []string{}
		found_by_slice []string = []string{}
	)

	if strings.Contains(this.net.Raw, "We're sorry, you are not allowed to proceed") {
		solve_slice = append(solve_slice, "One solution is to use proxy's")
		found_by_slice = append(found_by_slice, "Text snippet warning")
		confidence += 40
	} 
	
	if strings.Contains(this.net.Raw, "Your request looks suspicious or similar to automated requests from spam posting software") {
		solve_slice = append(solve_slice,"Set a time for requests with: --http-sleep")
		confidence += 40
	}

	defer this.wg.Done()

	this.model = &models.MiddlewareFirewallModel{Name: name, FoundBy: this.format_slice_to_string(found_by_slice), Solve: this.format_slice_to_string(solve_slice) , Confidence: confidence}
}

func (this *_waf) ninja_firewall() {
	var (
		name = "Ninja Firewall"
		confidence int
		solve_slice []string = []string{}
		found_by_slice []string = []string{}
	)

	if strings.Contains(this.net.Raw, "For security reasons, it was blocked and logged") {
		found_by_slice = append(found_by_slice, "Text snippet warning")
		solve_slice = append(solve_slice, "One solution is to use proxy's")
		confidence = 40
	} 
	
	if strings.Contains(this.net.Raw, "NinjaFirewall") && strings.Contains(this.net.Raw, "NinjaFirewall: 403 Forbidden") {
		found_by_slice = append(found_by_slice, "Keyword in title.")
		confidence = 50
	}

	defer this.wg.Done()

	this.model = &models.MiddlewareFirewallModel{Name: name, FoundBy: this.format_slice_to_string(found_by_slice), Solve: this.format_slice_to_string(solve_slice) , Confidence: confidence}
}

func (this *_waf) wordfence() {
	var (
		name = "Wordfence"
		confidence int
		found_by_slice []string = []string{}
		solve_slice []string = []string{}
	)
	
	if strings.Contains(this.net.Raw, "Generated by Wordfence") || strings.Contains(this.net.Raw, "This response was generated by Wordfence") {
		found_by_slice = append(found_by_slice, "Text snippet")
		confidence = 80
	}
	
	if strings.Contains(this.net.Raw, "A potentially unsafe operation has been detected in your request to this site") || strings.Contains(this.net.Raw, "Your access to this site has been limited") {
		found_by_slice = append(found_by_slice, "Text snippet warning")
		solve_slice = append(solve_slice, "Set a time for requests with: --http-sleep")
		confidence = 20
	}

	defer this.wg.Done()

	this.model = &models.MiddlewareFirewallModel{Name: name, FoundBy: this.format_slice_to_string(found_by_slice), Solve: this.format_slice_to_string(solve_slice) , Confidence: confidence}
}

func (this *_waf) bullet_proof() {
	var (
		name = "BulletProof Security"
		confidence int
		found_by_slice []string = []string{}
	)

	if strings.Contains(this.net.Raw, "If you arrived here due to a search or clicking on a link click your Browser's back button to return to the previous page.") {
		found_by_slice = append(found_by_slice, "Text snippet")
		confidence = 10
	}

	defer this.wg.Done()

	this.model = &models.MiddlewareFirewallModel{Name: name, FoundBy: this.format_slice_to_string(found_by_slice) , Confidence: confidence}
}

func (this *_waf) site_guard() {
	var (
		name = "SiteGuard"
		confidence int
		found_by_slice []string = []string{}
	)

	if strings.Contains(this.net.Raw, "Powered by SiteGuard") || strings.Contains(this.net.Raw, "The server refuse to browse the page.") {
		found_by_slice = append(found_by_slice, "Text snippet")
		confidence = 20
	}

	defer this.wg.Done()

	this.model = &models.MiddlewareFirewallModel{Name: name, FoundBy: this.format_slice_to_string(found_by_slice), Confidence: confidence}
}

func (this *_waf) all() *models.MiddlewareFirewallModel {
	this.wg.Add(6)

	go this.cloudflare()
	go this.site_guard()
	go this.bullet_proof()
	go this.cerber()
	go this.wordfence()
	go this.ninja_firewall()

	this.wg.Wait()

	return this.model
}