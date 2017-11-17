package slack

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var simpleMessage = `{
    "type": "message",
    "channel": "C2147483705",
    "user": "U2147483697",
    "text": "Hello world",
    "ts": "1355517523.000005"
}`

func unmarshalMessage(j string) (*Message, error) {
	message := &Message{}
	if err := json.Unmarshal([]byte(j), &message); err != nil {
		return nil, err
	}
	return message, nil
}

func TestSimpleMessage(t *testing.T) {
	message, err := unmarshalMessage(simpleMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "C2147483705", message.Channel)
	assert.Equal(t, "U2147483697", message.User)
	assert.Equal(t, "Hello world", message.Text)
	assert.Equal(t, "1355517523.000005", message.Timestamp)
}

var starredMessage = `{
    "text": "is testing",
    "type": "message",
    "subtype": "me_message",
    "user": "U2147483697",
    "ts": "1433314126.000003",
    "is_starred": true
}`

func TestStarredMessage(t *testing.T) {
	message, err := unmarshalMessage(starredMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "is testing", message.Text)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "me_message", message.SubType)
	assert.Equal(t, "U2147483697", message.User)
	assert.Equal(t, "1433314126.000003", message.Timestamp)
	assert.Equal(t, true, message.IsStarred)
}

var editedMessage = `{
    "type": "message",
    "user": "U2147483697",
    "text": "hello edited",
    "edited": {
        "user": "U2147483697",
        "ts": "1433314416.000000"
    },
    "ts": "1433314408.000004"
}`

func TestEditedMessage(t *testing.T) {
	message, err := unmarshalMessage(editedMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "U2147483697", message.User)
	assert.Equal(t, "hello edited", message.Text)
	assert.NotNil(t, message.Edited)
	assert.Equal(t, "U2147483697", message.Edited.User)
	assert.Equal(t, "1433314416.000000", message.Edited.Timestamp)
	assert.Equal(t, "1433314408.000004", message.Timestamp)
}

var uploadedFile = `{
    "type": "message",
    "subtype": "file_share",
    "text": "<@U2147483697|tester> uploaded a file: <https:\/\/test.slack.com\/files\/tester\/abc\/test.txt|test.txt> and commented: test comment here",
    "file": {
        "id": "abc",
        "created": 1433314757,
        "timestamp": 1433314757,
        "name": "test.txt",
        "title": "test.txt",
        "mimetype": "text\/plain",
        "filetype": "text",
        "pretty_type": "Plain Text",
        "user": "U2147483697",
        "editable": true,
        "size": 5,
        "mode": "snippet",
        "is_external": false,
        "external_type": "",
        "is_public": true,
        "public_url_shared": false,
        "url": "https:\/\/slack-files.com\/files-pub\/abc-def-ghi\/test.txt",
        "url_download": "https:\/\/slack-files.com\/files-pub\/abc-def-ghi\/download\/test.txt",
        "url_private": "https:\/\/files.slack.com\/files-pri\/abc-def\/test.txt",
        "url_private_download": "https:\/\/files.slack.com\/files-pri\/abc-def\/download\/test.txt",
        "permalink": "https:\/\/test.slack.com\/files\/tester\/abc\/test.txt",
        "permalink_public": "https:\/\/slack-files.com\/abc-def-ghi",
        "edit_link": "https:\/\/test.slack.com\/files\/tester\/abc\/test.txt\/edit",
        "preview": "test\n",
        "preview_highlight": "<div class=\"sssh-code\"><div class=\"sssh-line\"><pre>test<\/pre><\/div>\n<div class=\"sssh-line\"><pre><\/pre><\/div>\n<\/div>",
        "lines": 2,
        "lines_more": 0,
        "channels": [
            "C2147483705"
        ],
        "groups": [],
        "ims": [],
        "comments_count": 1,
        "initial_comment": {
            "id": "Fc066YLGKH",
            "created": 1433314757,
            "timestamp": 1433314757,
            "user": "U2147483697",
            "comment": "test comment here"
        }
    },
    "user": "U2147483697",
    "upload": true,
    "ts": "1433314757.000006"
}`

func TestUploadedFile(t *testing.T) {
	message, err := unmarshalMessage(uploadedFile)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "file_share", message.SubType)
	assert.Equal(t, "<@U2147483697|tester> uploaded a file: <https://test.slack.com/files/tester/abc/test.txt|test.txt> and commented: test comment here", message.Text)
	// TODO: Assert File
	assert.Equal(t, "U2147483697", message.User)
	assert.True(t, message.Upload)
	assert.Equal(t, "1433314757.000006", message.Timestamp)
}

var testPost = `{
    "type": "message",
    "subtype": "file_share",
    "text": "<@U2147483697|tester> shared a file: <https:\/\/test.slack.com\/files\/tester\/abc\/test_post|test post>",
    "file": {
        "id": "abc",
        "created": 1433315398,
        "timestamp": 1433315398,
        "name": "test_post",
        "title": "test post",
        "mimetype": "text\/plain",
        "filetype": "post",
        "pretty_type": "Post",
        "user": "U2147483697",
        "editable": true,
        "size": 14,
        "mode": "post",
        "is_external": false,
        "external_type": "",
        "is_public": true,
        "public_url_shared": false,
        "url": "https:\/\/slack-files.com\/files-pub\/abc-def-ghi\/test_post",
        "url_download": "https:\/\/slack-files.com\/files-pub\/abc-def-ghi\/download\/test_post",
        "url_private": "https:\/\/files.slack.com\/files-pri\/abc-def\/test_post",
        "url_private_download": "https:\/\/files.slack.com\/files-pri\/abc-def\/download\/test_post",
        "permalink": "https:\/\/test.slack.com\/files\/tester\/abc\/test_post",
        "permalink_public": "https:\/\/slack-files.com\/abc-def-ghi",
        "edit_link": "https:\/\/test.slack.com\/files\/tester\/abc\/test_post\/edit",
        "preview": "test post body",
        "channels": [
            "C2147483705"
        ],
        "groups": [],
        "ims": [],
        "comments_count": 1
    },
    "user": "U2147483697",
    "upload": false,
    "ts": "1433315416.000008"
}`

func TestPost(t *testing.T) {
	message, err := unmarshalMessage(testPost)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "file_share", message.SubType)
	assert.Equal(t, "<@U2147483697|tester> shared a file: <https://test.slack.com/files/tester/abc/test_post|test post>", message.Text)
	// TODO: Assert File
	assert.Equal(t, "U2147483697", message.User)
	assert.False(t, message.Upload)
	assert.Equal(t, "1433315416.000008", message.Timestamp)
}

var testComment = `{
    "type": "message",
    "subtype": "file_comment",
    "text": "<@U2147483697|tester> commented on <@U2147483697|tester>'s file <https:\/\/test.slack.com\/files\/tester\/abc\/test_post|test post>: another comment",
    "file": {
        "id": "abc",
        "created": 1433315398,
        "timestamp": 1433315398,
        "name": "test_post",
        "title": "test post",
        "mimetype": "text\/plain",
        "filetype": "post",
        "pretty_type": "Post",
        "user": "U2147483697",
        "editable": true,
        "size": 14,
        "mode": "post",
        "is_external": false,
        "external_type": "",
        "is_public": true,
        "public_url_shared": false,
        "url": "https:\/\/slack-files.com\/files-pub\/abc-def-ghi\/test_post",
        "url_download": "https:\/\/slack-files.com\/files-pub\/abc-def-ghi\/download\/test_post",
        "url_private": "https:\/\/files.slack.com\/files-pri\/abc-def\/test_post",
        "url_private_download": "https:\/\/files.slack.com\/files-pri\/abc-def\/download\/test_post",
        "permalink": "https:\/\/test.slack.com\/files\/tester\/abc\/test_post",
        "permalink_public": "https:\/\/slack-files.com\/abc-def-ghi",
        "edit_link": "https:\/\/test.slack.com\/files\/tester\/abc\/test_post\/edit",
        "preview": "test post body",
        "channels": [
            "C2147483705"
        ],
        "groups": [],
        "ims": [],
        "comments_count": 2
    },
    "comment": {
        "id": "xyz",
        "created": 1433316360,
        "timestamp": 1433316360,
        "user": "U2147483697",
        "comment": "another comment"
    },
    "ts": "1433316360.000009"
}`

func TestComment(t *testing.T) {
	message, err := unmarshalMessage(testComment)
	fmt.Println(err)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "file_comment", message.SubType)
	assert.Equal(t, "<@U2147483697|tester> commented on <@U2147483697|tester>'s file <https://test.slack.com/files/tester/abc/test_post|test post>: another comment", message.Text)
	// TODO: Assert File
	// TODO: Assert Comment
	assert.Equal(t, "1433316360.000009", message.Timestamp)
}

var botMessage = `{
    "type": "message",
    "subtype": "bot_message",
    "ts": "1358877455.000010",
    "text": "Pushing is the answer",
    "bot_id": "BB12033",
    "username": "github",
    "icons": {}
}`

func TestBotMessage(t *testing.T) {
	message, err := unmarshalMessage(botMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "bot_message", message.SubType)
	assert.Equal(t, "1358877455.000010", message.Timestamp)
	assert.Equal(t, "Pushing is the answer", message.Text)
	assert.Equal(t, "BB12033", message.BotID)
	assert.Equal(t, "github", message.Username)
	assert.NotNil(t, message.Icons)
	assert.Empty(t, message.Icons.IconURL)
	assert.Empty(t, message.Icons.IconEmoji)
}

var meMessage = `{
    "type": "message",
    "subtype": "me_message",
    "channel": "C2147483705",
    "user": "U2147483697",
    "text": "is doing that thing",
    "ts": "1355517523.000005"
}`

func TestMeMessage(t *testing.T) {
	message, err := unmarshalMessage(meMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "me_message", message.SubType)
	assert.Equal(t, "C2147483705", message.Channel)
	assert.Equal(t, "U2147483697", message.User)
	assert.Equal(t, "is doing that thing", message.Text)
	assert.Equal(t, "1355517523.000005", message.Timestamp)
}

var messageChangedMessage = `{
    "type": "message",
    "subtype": "message_changed",
    "hidden": true,
    "channel": "C2147483705",
    "ts": "1358878755.000001",
    "message": {
        "type": "message",
        "user": "U2147483697",
        "text": "Hello, world!",
        "ts": "1355517523.000005",
        "edited": {
            "user": "U2147483697",
            "ts": "1358878755.000001"
        }
    }
}`

func TestMessageChangedMessage(t *testing.T) {
	message, err := unmarshalMessage(messageChangedMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "message_changed", message.SubType)
	assert.True(t, message.Hidden)
	assert.Equal(t, "C2147483705", message.Channel)
	assert.NotNil(t, message.SubMessage)
	assert.Equal(t, "message", message.SubMessage.Type)
	assert.Equal(t, "U2147483697", message.SubMessage.User)
	assert.Equal(t, "Hello, world!", message.SubMessage.Text)
	assert.Equal(t, "1355517523.000005", message.SubMessage.Timestamp)
	assert.NotNil(t, message.SubMessage.Edited)
	assert.Equal(t, "U2147483697", message.SubMessage.Edited.User)
	assert.Equal(t, "1358878755.000001", message.SubMessage.Edited.Timestamp)
	assert.Equal(t, "1358878755.000001", message.Timestamp)
}

var messageDeletedMessage = `{
    "type": "message",
    "subtype": "message_deleted",
    "hidden": true,
    "channel": "C2147483705",
    "ts": "1358878755.000001",
    "deleted_ts": "1358878749.000002"
}`

func TestMessageDeletedMessage(t *testing.T) {
	message, err := unmarshalMessage(messageDeletedMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "message_deleted", message.SubType)
	assert.True(t, message.Hidden)
	assert.Equal(t, "C2147483705", message.Channel)
	assert.Equal(t, "1358878755.000001", message.Timestamp)
	assert.Equal(t, "1358878749.000002", message.DeletedTimestamp)
}

var channelJoinMessage = `{
    "type": "message",
    "subtype": "channel_join",
    "ts": "1358877458.000011",
    "user": "U2147483828",
    "text": "<@U2147483828|cal> has joined the channel"
}`

func TestChannelJoinMessage(t *testing.T) {
	message, err := unmarshalMessage(channelJoinMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "channel_join", message.SubType)
	assert.Equal(t, "1358877458.000011", message.Timestamp)
	assert.Equal(t, "U2147483828", message.User)
	assert.Equal(t, "<@U2147483828|cal> has joined the channel", message.Text)
}

var channelJoinInvitedMessage = `{
    "type": "message",
    "subtype": "channel_join",
    "ts": "1358877458.000011",
    "user": "U2147483828",
    "text": "<@U2147483828|cal> has joined the channel",
		"inviter": "U2147483829"
}`

func TestChannelJoinInvitedMessage(t *testing.T) {
	message, err := unmarshalMessage(channelJoinInvitedMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "channel_join", message.SubType)
	assert.Equal(t, "1358877458.000011", message.Timestamp)
	assert.Equal(t, "U2147483828", message.User)
	assert.Equal(t, "<@U2147483828|cal> has joined the channel", message.Text)
	assert.Equal(t, "U2147483829", message.Inviter)
}

var channelLeaveMessage = `{
    "type": "message",
    "subtype": "channel_leave",
    "ts": "1358877455.000010",
    "user": "U2147483828",
    "text": "<@U2147483828|cal> has left the channel"
}`

func TestChannelLeaveMessage(t *testing.T) {
	message, err := unmarshalMessage(channelLeaveMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "channel_leave", message.SubType)
	assert.Equal(t, "1358877455.000010", message.Timestamp)
	assert.Equal(t, "U2147483828", message.User)
	assert.Equal(t, "<@U2147483828|cal> has left the channel", message.Text)
}

var channelTopicMessage = `{
    "type": "message",
    "subtype": "channel_topic",
    "ts": "1358877455.000010",
    "user": "U2147483828",
    "topic": "hello world",
    "text": "<@U2147483828|cal> set the channel topic: hello world"
}`

func TestChannelTopicMessage(t *testing.T) {
	message, err := unmarshalMessage(channelTopicMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "channel_topic", message.SubType)
	assert.Equal(t, "1358877455.000010", message.Timestamp)
	assert.Equal(t, "U2147483828", message.User)
	assert.Equal(t, "hello world", message.Topic)
	assert.Equal(t, "<@U2147483828|cal> set the channel topic: hello world", message.Text)
}

var channelPurposeMessage = `{
    "type": "message",
    "subtype": "channel_purpose",
    "ts": "1358877455.000010",
    "user": "U2147483828",
    "purpose": "whatever",
    "text": "<@U2147483828|cal> set the channel purpose: whatever"
}`

func TestChannelPurposeMessage(t *testing.T) {
	message, err := unmarshalMessage(channelPurposeMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "channel_purpose", message.SubType)
	assert.Equal(t, "1358877455.000010", message.Timestamp)
	assert.Equal(t, "U2147483828", message.User)
	assert.Equal(t, "whatever", message.Purpose)
	assert.Equal(t, "<@U2147483828|cal> set the channel purpose: whatever", message.Text)
}

var channelNameMessage = `{
    "type": "message",
    "subtype": "channel_name",
    "ts": "1358877455.000010",
    "user": "U2147483828",
    "old_name": "random",
    "name": "watercooler",
    "text": "<@U2147483828|cal> has renamed the channel from \"random\" to \"watercooler\""
}`

func TestChannelNameMessage(t *testing.T) {
	message, err := unmarshalMessage(channelNameMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "channel_name", message.SubType)
	assert.Equal(t, "1358877455.000010", message.Timestamp)
	assert.Equal(t, "U2147483828", message.User)
	assert.Equal(t, "random", message.OldName)
	assert.Equal(t, "watercooler", message.Name)
	assert.Equal(t, "<@U2147483828|cal> has renamed the channel from \"random\" to \"watercooler\"", message.Text)
}

var channelArchiveMessage = `{
    "type": "message",
    "subtype": "channel_archive",
    "ts": "1361482916.000003",
    "text": "<U1234|@cal> archived the channel",
    "user": "U1234",
    "members": ["U1234", "U5678"]
}`

func TestChannelArchiveMessage(t *testing.T) {
	message, err := unmarshalMessage(channelArchiveMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "channel_archive", message.SubType)
	assert.Equal(t, "1361482916.000003", message.Timestamp)
	assert.Equal(t, "<U1234|@cal> archived the channel", message.Text)
	assert.Equal(t, "U1234", message.User)
	assert.NotNil(t, message.Members)
	assert.Equal(t, 2, len(message.Members))
}

var channelUnarchiveMessage = `{
    "type": "message",
    "subtype": "channel_unarchive",
    "ts": "1361482916.000003",
    "text": "<U1234|@cal> un-archived the channel",
    "user": "U1234"
}`

func TestChannelUnarchiveMessage(t *testing.T) {
	message, err := unmarshalMessage(channelUnarchiveMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "channel_unarchive", message.SubType)
	assert.Equal(t, "1361482916.000003", message.Timestamp)
	assert.Equal(t, "<U1234|@cal> un-archived the channel", message.Text)
	assert.Equal(t, "U1234", message.User)
}

var channelRepliesParentMessage = `{
    "type": "message",
    "user": "U1234",
    "text": "test",
    "thread_ts": "1493305433.915644",
    "reply_count": 2,
    "replies": [
        {
            "user": "U5678",
            "ts": "1493305444.920992"
        },
        {
            "user": "U9012",
            "ts": "1493305894.133936"
        }
    ],
    "subscribed": true,
    "last_read": "1493305894.133936",
    "unread_count": 0,
    "ts": "1493305433.915644"
}`

func TestChannelRepliesParentMessage(t *testing.T) {
	message, err := unmarshalMessage(channelRepliesParentMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "U1234", message.User)
	assert.Equal(t, "test", message.Text)
	assert.Equal(t, "1493305433.915644", message.ThreadTimestamp)
	assert.Equal(t, 2, message.ReplyCount)
	assert.Equal(t, "U5678", message.Replies[0].User)
	assert.Equal(t, "1493305444.920992", message.Replies[0].Timestamp)
	assert.Equal(t, "U9012", message.Replies[1].User)
	assert.Equal(t, "1493305894.133936", message.Replies[1].Timestamp)
	assert.Equal(t, "1493305433.915644", message.Timestamp)
}

var channelRepliesChildMessage = `{
    "type": "message",
    "user": "U5678",
    "text": "foo",
    "thread_ts": "1493305433.915644",
    "parent_user_id": "U1234",
    "ts": "1493305444.920992"
}`

func TestChannelRepliesChildMessage(t *testing.T) {
	message, err := unmarshalMessage(channelRepliesChildMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "U5678", message.User)
	assert.Equal(t, "foo", message.Text)
	assert.Equal(t, "1493305433.915644", message.ThreadTimestamp)
	assert.Equal(t, "U1234", message.ParentUserId)
	assert.Equal(t, "1493305444.920992", message.Timestamp)
}

var groupJoinMessage = `{
    "type": "message",
    "subtype": "group_join",
    "ts": "1358877458.000011",
    "user": "U2147483828",
    "text": "<@U2147483828|cal> has joined the group"
}`

func TestGroupJoinMessage(t *testing.T) {
	message, err := unmarshalMessage(groupJoinMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "group_join", message.SubType)
	assert.Equal(t, "1358877458.000011", message.Timestamp)
	assert.Equal(t, "U2147483828", message.User)
	assert.Equal(t, "<@U2147483828|cal> has joined the group", message.Text)
}

var groupJoinInvitedMessage = `{
    "type": "message",
    "subtype": "group_join",
    "ts": "1358877458.000011",
    "user": "U2147483828",
    "text": "<@U2147483828|cal> has joined the group",
		"inviter": "U2147483829"
}`

func TestGroupJoinInvitedMessage(t *testing.T) {
	message, err := unmarshalMessage(groupJoinInvitedMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "group_join", message.SubType)
	assert.Equal(t, "1358877458.000011", message.Timestamp)
	assert.Equal(t, "U2147483828", message.User)
	assert.Equal(t, "<@U2147483828|cal> has joined the group", message.Text)
	assert.Equal(t, "U2147483829", message.Inviter)
}

var groupLeaveMessage = `{
    "type": "message",
    "subtype": "group_leave",
    "ts": "1358877455.000010",
    "user": "U2147483828",
    "text": "<@U2147483828|cal> has left the group"
}`

func TestGroupLeaveMessage(t *testing.T) {
	message, err := unmarshalMessage(groupLeaveMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "group_leave", message.SubType)
	assert.Equal(t, "1358877455.000010", message.Timestamp)
	assert.Equal(t, "U2147483828", message.User)
	assert.Equal(t, "<@U2147483828|cal> has left the group", message.Text)
}

var groupTopicMessage = `{
    "type": "message",
    "subtype": "group_topic",
    "ts": "1358877455.000010",
    "user": "U2147483828",
    "topic": "hello world",
    "text": "<@U2147483828|cal> set the group topic: hello world"
}`

func TestGroupTopicMessage(t *testing.T) {
	message, err := unmarshalMessage(groupTopicMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "group_topic", message.SubType)
	assert.Equal(t, "1358877455.000010", message.Timestamp)
	assert.Equal(t, "U2147483828", message.User)
	assert.Equal(t, "hello world", message.Topic)
	assert.Equal(t, "<@U2147483828|cal> set the group topic: hello world", message.Text)
}

var groupPurposeMessage = `{
    "type": "message",
    "subtype": "group_purpose",
    "ts": "1358877455.000010",
    "user": "U2147483828",
    "purpose": "whatever",
    "text": "<@U2147483828|cal> set the group purpose: whatever"
}`

func TestGroupPurposeMessage(t *testing.T) {
	message, err := unmarshalMessage(groupPurposeMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "group_purpose", message.SubType)
	assert.Equal(t, "1358877455.000010", message.Timestamp)
	assert.Equal(t, "U2147483828", message.User)
	assert.Equal(t, "whatever", message.Purpose)
	assert.Equal(t, "<@U2147483828|cal> set the group purpose: whatever", message.Text)
}

var groupNameMessage = `{
    "type": "message",
    "subtype": "group_name",
    "ts": "1358877455.000010",
    "user": "U2147483828",
    "old_name": "random",
    "name": "watercooler",
    "text": "<@U2147483828|cal> has renamed the group from \"random\" to \"watercooler\""
}`

func TestGroupNameMessage(t *testing.T) {
	message, err := unmarshalMessage(groupNameMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "group_name", message.SubType)
	assert.Equal(t, "1358877455.000010", message.Timestamp)
	assert.Equal(t, "U2147483828", message.User)
	assert.Equal(t, "random", message.OldName)
	assert.Equal(t, "watercooler", message.Name)
	assert.Equal(t, "<@U2147483828|cal> has renamed the group from \"random\" to \"watercooler\"", message.Text)
}

var groupArchiveMessage = `{
    "type": "message",
    "subtype": "group_archive",
    "ts": "1361482916.000003",
    "text": "<U1234|@cal> archived the group",
    "user": "U1234",
    "members": ["U1234", "U5678"]
}`

func TestGroupArchiveMessage(t *testing.T) {
	message, err := unmarshalMessage(groupArchiveMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "group_archive", message.SubType)
	assert.Equal(t, "1361482916.000003", message.Timestamp)
	assert.Equal(t, "<U1234|@cal> archived the group", message.Text)
	assert.Equal(t, "U1234", message.User)
	assert.NotNil(t, message.Members)
	assert.Equal(t, 2, len(message.Members))
}

var groupUnarchiveMessage = `{
    "type": "message",
    "subtype": "group_unarchive",
    "ts": "1361482916.000003",
    "text": "<U1234|@cal> un-archived the group",
    "user": "U1234"
}`

func TestGroupUnarchiveMessage(t *testing.T) {
	message, err := unmarshalMessage(groupUnarchiveMessage)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "group_unarchive", message.SubType)
	assert.Equal(t, "1361482916.000003", message.Timestamp)
	assert.Equal(t, "<U1234|@cal> un-archived the group", message.Text)
	assert.Equal(t, "U1234", message.User)
}

var fileShareMessage = `{
    "type": "message",
    "subtype": "file_share",
    "ts": "1358877455.000010",
    "text": "<@cal> uploaded a file: <https:...7.png|7.png>",
    "file": {
        "id" : "F2147483862",
        "created" : 1356032811,
        "timestamp" : 1356032811,
        "name" : "file.htm",
        "title" : "My HTML file",
        "mimetype" : "text\/plain",
        "filetype" : "text",
        "pretty_type": "Text",
        "user" : "U2147483697",
        "mode" : "hosted",
        "editable" : true,
        "is_external": false,
        "external_type": "",
        "size" : 12345,
        "url": "https:\/\/slack-files.com\/files-pub\/T024BE7LD-F024BERPE-09acb6\/1.png",
        "url_download": "https:\/\/slack-files.com\/files-pub\/T024BE7LD-F024BERPE-09acb6\/download\/1.png",
        "url_private": "https:\/\/slack.com\/files-pri\/T024BE7LD-F024BERPE\/1.png",
        "url_private_download": "https:\/\/slack.com\/files-pri\/T024BE7LD-F024BERPE\/download\/1.png",
        "thumb_64": "https:\/\/slack-files.com\/files-tmb\/T024BE7LD-F024BERPE-c66246\/1_64.png",
        "thumb_80": "https:\/\/slack-files.com\/files-tmb\/T024BE7LD-F024BERPE-c66246\/1_80.png",
        "thumb_360": "https:\/\/slack-files.com\/files-tmb\/T024BE7LD-F024BERPE-c66246\/1_360.png",
        "thumb_360_gif": "https:\/\/slack-files.com\/files-tmb\/T024BE7LD-F024BERPE-c66246\/1_360.gif",
        "thumb_360_w": 100,
        "thumb_360_h": 100,
        "permalink" : "https:\/\/tinyspeck.slack.com\/files\/cal\/F024BERPE\/1.png",
        "edit_link" : "https:\/\/tinyspeck.slack.com\/files\/cal\/F024BERPE\/1.png/edit",
        "preview" : "&lt;!DOCTYPE html&gt;\n&lt;html&gt;\n&lt;meta charset='utf-8'&gt;",
        "preview_highlight" : "&lt;div class=\"sssh-code\"&gt;&lt;div class=\"sssh-line\"&gt;&lt;pre&gt;&lt;!DOCTYPE html...",
        "lines" : 123,
        "lines_more": 118,
        "is_public": true,
        "public_url_shared": false,
        "channels": ["C024BE7LT"],
        "groups": ["G12345"],
        "ims": ["D12345"],
        "initial_comment": {},
        "num_stars": 7,
        "is_starred": true
    },
    "user": "U2147483697",
    "upload": true
}`

func TestFileShareMessage(t *testing.T) {
	message, err := unmarshalMessage(fileShareMessage)
	fmt.Println(err)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "file_share", message.SubType)
	assert.Equal(t, "1358877455.000010", message.Timestamp)
	assert.Equal(t, "<@cal> uploaded a file: <https:...7.png|7.png>", message.Text)
	assert.Equal(t, "U2147483697", message.User)
	assert.True(t, message.Upload)
	assert.NotNil(t, message.File)
}
