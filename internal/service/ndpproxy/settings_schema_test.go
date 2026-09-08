package ndpproxy

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	apindp "github.com/biptec/opnsense-go/pkg/ndpproxy"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSettingsRoundTrip(t *testing.T) {
	remote := &apindp.SettingsResponse{NdpProxy: apindp.Settings{General: apindp.GeneralSettings{
		Enabled: "1", Upstream: api.SelectedMap("wan"), Downstream: api.SelectedMapList{"opt2", "opt1"},
		RA: "0", Routes: "0", CarpDependOn: "1", Debug: "0",
	}}}
	model := settingsAPIToModel(remote)
	if model.Upstream.ValueString() != "wan" || !model.Enabled.ValueBool() || model.InstallHostRoutes.ValueBool() || !model.CARPDependOn.ValueBool() {
		t.Fatalf("unexpected model: %+v", model)
	}

	plan := &settingsResourceModel{
		Enabled: types.BoolValue(true), Upstream: types.StringValue("wan"),
		Downstream:                tools.StringSliceToSet([]string{"opt1", "opt2"}),
		ProxyRouterAdvertisements: types.BoolValue(false), InstallHostRoutes: types.BoolValue(false),
		CARPDependOn: types.BoolValue(true), Debug: types.BoolValue(false),
	}
	applySettingsModel(&remote.NdpProxy.General, plan)
	g := remote.NdpProxy.General
	if g.Upstream.String() != "wan" || g.Downstream.String() != "opt1,opt2" || g.RA != "0" || g.Routes != "0" || g.CarpDependOn != "1" {
		t.Fatalf("unexpected API model: %+v", g)
	}
}
