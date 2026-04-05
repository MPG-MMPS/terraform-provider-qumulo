package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = (*qumuloProvider)(nil)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &qumuloProvider{}
	}
}

type qumuloProvider struct{}

func (p *qumuloProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {

}

func (p *qumuloProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

}

func (p *qumuloProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "qumulo"
}

func (p *qumuloProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *qumuloProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFilesQuotasResource,
	}
}
