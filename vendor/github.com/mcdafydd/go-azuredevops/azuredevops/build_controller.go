package azuredevops

// BuildController represents a controller of the build service
type BuildController struct {
	CreatedDate *string `json:"createdDate"`
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
	ID          *int    `json:"id"`
	Name        *string `json:"name"`
	Status      *string `json:"status"`
	UpdateDate  *string `json:"updateDate"`
	URI         *string `json:"uri"`
	URL         *string `json:"url"`
}
