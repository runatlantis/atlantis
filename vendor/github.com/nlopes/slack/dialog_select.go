package slack

// SelectDataSource types of select datasource
type SelectDataSource string

const (
	// DialogDataSourceStatic menu with static Options/OptionGroups
	DialogDataSourceStatic SelectDataSource = "static"
	// DialogDataSourceExternal dynamic datasource
	DialogDataSourceExternal SelectDataSource = "external"
	// DialogDataSourceConversations provides a list of conversations
	DialogDataSourceConversations SelectDataSource = "conversations"
	// DialogDataSourceChannels provides a list of channels
	DialogDataSourceChannels SelectDataSource = "channels"
	// DialogDataSourceUsers provides a list of users
	DialogDataSourceUsers SelectDataSource = "users"
)

// DialogInputSelect dialog support for select boxes.
type DialogInputSelect struct {
	DialogInput
	Value           string               `json:"value,omitempty"`            //Optional.
	DataSource      SelectDataSource     `json:"data_source,omitempty"`      //Optional. Allowed values: "users", "channels", "conversations", "external".
	SelectedOptions string               `json:"selected_options,omitempty"` //Optional. Default value for "external" only
	Options         []DialogSelectOption `json:"options,omitempty"`          //One of options or option_groups is required.
	OptionGroups    []DialogOptionGroup  `json:"option_groups,omitempty"`    //Provide up to 100 options.
}

// DialogSelectOption is an option for the user to select from the menu
type DialogSelectOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// DialogOptionGroup is a collection of options for creating a segmented table
type DialogOptionGroup struct {
	Label   string               `json:"label"`
	Options []DialogSelectOption `json:"options"`
}

// NewStaticSelectDialogInput constructor for a `static` datasource menu input
func NewStaticSelectDialogInput(name, label string, options []DialogSelectOption) *DialogInputSelect {
	return &DialogInputSelect{
		DialogInput: DialogInput{
			Type:     InputTypeSelect,
			Name:     name,
			Label:    label,
			Optional: true,
		},
		DataSource: DialogDataSourceStatic,
		Options:    options,
	}
}

// NewGroupedSelectDialogInput creates grouped options select input for Dialogs.
func NewGroupedSelectDialogInput(name, label string, groups map[string]map[string]string) *DialogInputSelect {
	optionGroups := []DialogOptionGroup{}
	for groupName, options := range groups {
		optionGroups = append(optionGroups, DialogOptionGroup{
			Label:   groupName,
			Options: optionsFromMap(options),
		})
	}
	return &DialogInputSelect{
		DialogInput: DialogInput{
			Type:  InputTypeSelect,
			Name:  name,
			Label: label,
		},
		DataSource:   DialogDataSourceStatic,
		OptionGroups: optionGroups,
	}
}

func optionsFromArray(options []string) []DialogSelectOption {
	selectOptions := make([]DialogSelectOption, len(options))
	for idx, value := range options {
		selectOptions[idx] = DialogSelectOption{
			Label: value,
			Value: value,
		}
	}
	return selectOptions
}

func optionsFromMap(options map[string]string) []DialogSelectOption {
	selectOptions := make([]DialogSelectOption, len(options))
	idx := 0
	var option DialogSelectOption
	for key, value := range options {
		option = DialogSelectOption{
			Label: key,
			Value: value,
		}
		selectOptions[idx] = option
		idx++
	}
	return selectOptions
}

// NewConversationsSelect returns a `Conversations` select
func NewConversationsSelect(name, label string) *DialogInputSelect {
	return newPresetSelect(name, label, DialogDataSourceConversations)
}

// NewChannelsSelect returns a `Channels` select
func NewChannelsSelect(name, label string) *DialogInputSelect {
	return newPresetSelect(name, label, DialogDataSourceChannels)
}

// NewUsersSelect returns a `Users` select
func NewUsersSelect(name, label string) *DialogInputSelect {
	return newPresetSelect(name, label, DialogDataSourceUsers)
}

func newPresetSelect(name, label string, dataSourceType SelectDataSource) *DialogInputSelect {
	return &DialogInputSelect{
		DialogInput: DialogInput{
			Type:  InputTypeSelect,
			Label: label,
			Name:  name,
		},
		DataSource: dataSourceType,
	}
}
