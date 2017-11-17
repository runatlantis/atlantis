package slack

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Channel
var simpleChannel = `{
    "id": "C024BE91L",
    "name": "fun",
    "is_channel": true,
    "created": 1360782804,
    "creator": "U024BE7LH",
    "is_archived": false,
    "is_general": false,
    "members": [
        "U024BE7LH"
    ],
    "topic": {
        "value": "Fun times",
        "creator": "U024BE7LV",
        "last_set": 1369677212
    },
    "purpose": {
        "value": "This channel is for fun",
        "creator": "U024BE7LH",
        "last_set": 1360782804
    },
    "is_member": true,
    "last_read": "1401383885.000061",
    "unread_count": 0,
    "unread_count_display": 0
}`

func unmarshalChannel(j string) (*Channel, error) {
	channel := &Channel{}
	if err := json.Unmarshal([]byte(j), &channel); err != nil {
		return nil, err
	}
	return channel, nil
}

func TestSimpleChannel(t *testing.T) {
	channel, err := unmarshalChannel(simpleChannel)
	assert.Nil(t, err)
	assertSimpleChannel(t, channel)
}

func assertSimpleChannel(t *testing.T, channel *Channel) {
	assert.NotNil(t, channel)
	assert.Equal(t, "C024BE91L", channel.ID)
	assert.Equal(t, "fun", channel.Name)
	assert.Equal(t, true, channel.IsChannel)
	assert.Equal(t, JSONTime(1360782804), channel.Created)
	assert.Equal(t, "U024BE7LH", channel.Creator)
	assert.Equal(t, false, channel.IsArchived)
	assert.Equal(t, false, channel.IsGeneral)
	assert.Equal(t, true, channel.IsMember)
	assert.Equal(t, "1401383885.000061", channel.LastRead)
	assert.Equal(t, 0, channel.UnreadCount)
	assert.Equal(t, 0, channel.UnreadCountDisplay)
}

func TestCreateSimpleChannel(t *testing.T) {
	channel := &Channel{}
	channel.ID = "C024BE91L"
	channel.Name = "fun"
	channel.IsChannel = true
	channel.Created = JSONTime(1360782804)
	channel.Creator = "U024BE7LH"
	channel.IsArchived = false
	channel.IsGeneral = false
	channel.IsMember = true
	channel.LastRead = "1401383885.000061"
	channel.UnreadCount = 0
	channel.UnreadCountDisplay = 0
	assertSimpleChannel(t, channel)
}

// Group
var simpleGroup = `{
    "id": "G024BE91L",
    "name": "secretplans",
    "is_group": true,
    "created": 1360782804,
    "creator": "U024BE7LH",
    "is_archived": false,
    "members": [
        "U024BE7LH"
    ],
    "topic": {
        "value": "Secret plans on hold",
        "creator": "U024BE7LV",
        "last_set": 1369677212
    },
    "purpose": {
        "value": "Discuss secret plans that no-one else should know",
        "creator": "U024BE7LH",
        "last_set": 1360782804
    },
    "last_read": "1401383885.000061",
    "unread_count": 0,
    "unread_count_display": 0
}`

func unmarshalGroup(j string) (*Group, error) {
	group := &Group{}
	if err := json.Unmarshal([]byte(j), &group); err != nil {
		return nil, err
	}
	return group, nil
}

func TestSimpleGroup(t *testing.T) {
	group, err := unmarshalGroup(simpleGroup)
	assert.Nil(t, err)
	assertSimpleGroup(t, group)
}

func assertSimpleGroup(t *testing.T, group *Group) {
	assert.NotNil(t, group)
	assert.Equal(t, "G024BE91L", group.ID)
	assert.Equal(t, "secretplans", group.Name)
	assert.Equal(t, true, group.IsGroup)
	assert.Equal(t, JSONTime(1360782804), group.Created)
	assert.Equal(t, "U024BE7LH", group.Creator)
	assert.Equal(t, false, group.IsArchived)
	assert.Equal(t, "1401383885.000061", group.LastRead)
	assert.Equal(t, 0, group.UnreadCount)
	assert.Equal(t, 0, group.UnreadCountDisplay)
}

func TestCreateSimpleGroup(t *testing.T) {
	group := &Group{}
	group.ID = "G024BE91L"
	group.Name = "secretplans"
	group.IsGroup = true
	group.Created = JSONTime(1360782804)
	group.Creator = "U024BE7LH"
	group.IsArchived = false
	group.LastRead = "1401383885.000061"
	group.UnreadCount = 0
	group.UnreadCountDisplay = 0
	assertSimpleGroup(t, group)
}

// IM
var simpleIM = `{
    "id": "D024BFF1M",
    "is_im": true,
    "user": "U024BE7LH",
    "created": 1360782804,
    "is_user_deleted": false,
    "is_open": true,
    "last_read": "1401383885.000061",
    "unread_count": 0,
    "unread_count_display": 0
}`

func unmarshalIM(j string) (*IM, error) {
	im := &IM{}
	if err := json.Unmarshal([]byte(j), &im); err != nil {
		return nil, err
	}
	return im, nil
}

func TestSimpleIM(t *testing.T) {
	im, err := unmarshalIM(simpleIM)
	assert.Nil(t, err)
	assertSimpleIM(t, im)
}

func assertSimpleIM(t *testing.T, im *IM) {
	assert.NotNil(t, im)
	assert.Equal(t, "D024BFF1M", im.ID)
	assert.Equal(t, true, im.IsIM)
	assert.Equal(t, JSONTime(1360782804), im.Created)
	assert.Equal(t, false, im.IsUserDeleted)
	assert.Equal(t, true, im.IsOpen)
	assert.Equal(t, "1401383885.000061", im.LastRead)
	assert.Equal(t, 0, im.UnreadCount)
	assert.Equal(t, 0, im.UnreadCountDisplay)
}

func TestCreateSimpleIM(t *testing.T) {
	im := &IM{}
	im.ID = "D024BFF1M"
	im.IsIM = true
	im.Created = JSONTime(1360782804)
	im.IsUserDeleted = false
	im.IsOpen = true
	im.LastRead = "1401383885.000061"
	im.UnreadCount = 0
	im.UnreadCountDisplay = 0
	assertSimpleIM(t, im)
}
