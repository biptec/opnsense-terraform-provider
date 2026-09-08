package ndpproxy

import (
	"sort"

	"github.com/biptec/opnsense-go/pkg/api"
	apindp "github.com/biptec/opnsense-go/pkg/ndpproxy"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type settingsResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Enabled                   types.Bool   `tfsdk:"enabled"`
	Upstream                  types.String `tfsdk:"upstream"`
	Downstream                types.Set    `tfsdk:"downstream"`
	ProxyRouterAdvertisements types.Bool   `tfsdk:"proxy_router_advertisements"`
	InstallHostRoutes         types.Bool   `tfsdk:"install_host_routes"`
	CARPDependOn              types.Bool   `tfsdk:"carp_depend_on"`
	Debug                     types.Bool   `tfsdk:"debug"`
}

func settingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages the built-in os-ndp-proxy-go singleton settings. Create adopts the existing plugin singleton. Routed endpoint deployments should disable proxy router advertisements and automatic host-route installation when Terraform already owns explicit /128 routes.",
		Attributes: map[string]schema.Attribute{
			"id":                          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"enabled":                     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"upstream":                    schema.StringAttribute{Required: true, MarkdownDescription: "Logical OPNsense interface used as the provider-facing NDP upstream, normally `wan`."},
			"downstream":                  schema.SetAttribute{Required: true, ElementType: types.StringType, Validators: []validator.Set{setvalidator.SizeAtLeast(1)}, MarkdownDescription: "Logical OPNsense endpoint interfaces on which proxied neighbors are reachable."},
			"proxy_router_advertisements": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"install_host_routes":         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether the proxy learns and installs host routes. Keep false when Terraform owns static /128 routes."},
			"carp_depend_on":              schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Run the proxy only while this router owns at least one CARP MASTER identity."},
			"debug":                       schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		},
	}
}

func settingsAPIToModel(remote *apindp.SettingsResponse) *settingsResourceModel {
	g := remote.NdpProxy.General
	return &settingsResourceModel{
		ID:                        types.StringValue("ndpproxy_settings"),
		Enabled:                   types.BoolValue(tools.StringToBool(g.Enabled)),
		Upstream:                  types.StringValue(g.Upstream.String()),
		Downstream:                tools.StringSliceToSet([]string(g.Downstream)),
		ProxyRouterAdvertisements: types.BoolValue(tools.StringToBool(g.RA)),
		InstallHostRoutes:         types.BoolValue(tools.StringToBool(g.Routes)),
		CARPDependOn:              types.BoolValue(tools.StringToBool(g.CarpDependOn)),
		Debug:                     types.BoolValue(tools.StringToBool(g.Debug)),
	}
}

func applySettingsModel(g *apindp.GeneralSettings, plan *settingsResourceModel) {
	g.Enabled = tools.BoolToString(plan.Enabled.ValueBool())
	g.Upstream = api.SelectedMap(plan.Upstream.ValueString())
	downstream := tools.SetToStringSlice(plan.Downstream)
	sort.Strings(downstream)
	g.Downstream = api.SelectedMapList(downstream)
	g.RA = tools.BoolToString(plan.ProxyRouterAdvertisements.ValueBool())
	g.Routes = tools.BoolToString(plan.InstallHostRoutes.ValueBool())
	g.CarpDependOn = tools.BoolToString(plan.CARPDependOn.ValueBool())
	g.Debug = tools.BoolToString(plan.Debug.ValueBool())
}
