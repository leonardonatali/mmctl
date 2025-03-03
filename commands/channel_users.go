// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mmctl/v6/client"
	"github.com/mattermost/mmctl/v6/printer"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var ChannelUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "Management of channel users",
}

var ChannelUsersAddCmd = &cobra.Command{
	Use:     "add [channel] [users]",
	Short:   "Add users to channel",
	Long:    "Add some users to channel",
	Example: "  channel users add myteam:mychannel user@example.com username",
	RunE:    withClient(channelUsersAddCmdF),
}

var ChannelUsersRemoveCmd = &cobra.Command{
	Use:   "remove [channel] [users]",
	Short: "Remove users from channel",
	Long:  "Remove some users from channel",
	Example: `  channel users remove myteam:mychannel user@example.com username
  channel users remove myteam:mychannel --all-users`,
	RunE: withClient(channelUsersRemoveCmdF),
}

func init() {
	ChannelUsersRemoveCmd.Flags().Bool("all-users", false, "Remove all users from the indicated channel.")

	ChannelUsersCmd.AddCommand(
		ChannelUsersAddCmd,
		ChannelUsersRemoveCmd,
	)

	ChannelCmd.AddCommand(ChannelUsersCmd)
}

func channelUsersAddCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return errors.New("not enough arguments")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.Errorf("unable to find channel %q", args[0])
	}

	users := getUsersFromUserArgs(c, args[1:])
	for i, user := range users {
		addUserToChannel(c, channel, user, args[i+1])
	}

	return nil
}

func addUserToChannel(c client.Client, channel *model.Channel, user *model.User, userArg string) {
	if user == nil {
		printer.PrintError("Can't find user '" + userArg + "'")
		return
	}
	if _, _, err := c.AddChannelMember(channel.Id, user.Id); err != nil {
		printer.PrintError("Unable to add '" + userArg + "' to " + channel.Name + ". Error: " + err.Error())
	}
}

func channelUsersRemoveCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	allUsers, _ := cmd.Flags().GetBool("all-users")

	if allUsers && len(args) != 1 {
		return errors.New("individual users must not be specified in conjunction with the --all-users flag")
	}

	if !allUsers && len(args) < 2 {
		return errors.New("you must specify some users to remove from the channel, or use the --all-users flag to remove them all")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.Errorf("unable to find channel %q", args[0])
	}

	if allUsers {
		removeAllUsersFromChannel(c, channel)
	} else {
		for i, user := range getUsersFromUserArgs(c, args[1:]) {
			if err := removeUserFromChannel(c, channel, user, args[i+1]); err != nil {
				return err
			}
		}
	}

	return nil
}

func removeUserFromChannel(c client.Client, channel *model.Channel, user *model.User, userArg string) error {
	if user == nil {
		return errors.Errorf("can't find user '%s'", userArg)
	}
	if _, err := c.RemoveUserFromChannel(channel.Id, user.Id); err != nil {
		return errors.Errorf("unable to remove '%s' from %s. Error: %s", userArg, channel.Name, err.Error())
	}

	return nil
}

func removeAllUsersFromChannel(c client.Client, channel *model.Channel) {
	members, _, err := c.GetChannelMembers(channel.Id, 0, 10000, "")
	if err != nil {
		printer.PrintError("Unable to remove all users from " + channel.Name + ". Error: " + err.Error())
	}

	for _, member := range members {
		if _, err := c.RemoveUserFromChannel(channel.Id, member.UserId); err != nil {
			printer.PrintError("Unable to remove '" + member.UserId + "' from " + channel.Name + ". Error: " + err.Error())
		}
	}
}
