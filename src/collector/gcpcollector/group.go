package gcpcollector

import (
	"fmt"
	"time"

	"github.com/nianticlabs/modron/src/pb"

	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (collector *GCPCollector) ListGroups(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	groups, err := collector.api.ListGroups(ctx)
	if err != nil {
		return nil, err
	}
	res := []*pb.Resource{}
	for _, group := range groups {
		creationDate, err := time.Parse(time.RFC3339, group.CreateTime)
		if err != nil {
			return nil, fmt.Errorf("parse timestamp %s of group %s: %v", group.CreateTime, group.Name, err)
		}
		updateDate, err := time.Parse(time.RFC3339, group.UpdateTime)
		if err != nil {
			return nil, fmt.Errorf("unable to parse update timestamp of group %q", group.Name)
		}
		key := &pb.IamGroup_EntityKey{
			Id:        group.GroupKey.Id,
			Namespace: group.GroupKey.Namespace,
		}
		members := []*pb.IamGroup_Member{}
		groupMembers, err := collector.api.ListGroupMembers(ctx, group.Name)
		if err != nil {
			return nil, fmt.Errorf("unable to get members of group %q", group.Name)
		}
		for _, member := range groupMembers {
			key := pb.IamGroup_EntityKey{
				Id:        member.PreferredMemberKey.Id,
				Namespace: member.PreferredMemberKey.Namespace,
			}
			var memberRole pb.IamGroup_Member_Role
			switch member.Roles[0].Name {
			case "OWNER":
				memberRole = pb.IamGroup_Member_MEMBER_ROLE_OWNER
			case "MANAGER":
				memberRole = pb.IamGroup_Member_MEMBER_ROLE_MANAGER
			case "MEMBER":
				memberRole = pb.IamGroup_Member_MEMBER_ROLE_MEMBER
			default:
				memberRole = pb.IamGroup_Member_MEMBER_ROLE_UNKNOWN
			}
			var memberType pb.IamGroup_Member_Type
			switch member.Type {
			case "USER":
				memberType = pb.IamGroup_Member_MEMBER_TYPE_USER
			case "GROUP":
				memberType = pb.IamGroup_Member_MEMBER_TYPE_GROUP
			case "SERVICE_ACCOUNT":
				memberType = pb.IamGroup_Member_MEMBER_TYPE_SERVICE_ACCOUNT
			case "SHARED_DRIVE":
				memberType = pb.IamGroup_Member_MEMBER_TYPE_SHARED_DRIVE
			case "OTHER":
				memberType = pb.IamGroup_Member_MEMBER_TYPE_OTHER
			default:
				memberType = pb.IamGroup_Member_MEMBER_TYPE_UNKNOWN
			}
			joinDate, err := time.Parse(time.RFC3339, member.CreateTime)
			if err != nil {
				return nil, fmt.Errorf("unable to parse join timestamp of member %q", member.Name)
			}
			groupMember := pb.IamGroup_Member{
				Key:      &key,
				Role:     memberRole,
				Type:     memberType,
				JoinDate: timestamppb.New(joinDate),
			}
			members = append(members, &groupMember)
		}

		groupResource := &pb.IamGroup{
			Name:         group.Name,
			DisplayName:  group.DisplayName,
			Description:  group.Description,
			Key:          key,
			Parent:       group.Parent,
			CreationDate: timestamppb.New(creationDate),
			UpdateDate:   timestamppb.New(updateDate),
			Member:       members,
		}

		resource := &pb.Resource{
			Name:              groupResource.Name,
			Parent:            groupResource.Parent,
			ResourceGroupName: groupResource.Parent,
			Type: &pb.Resource_Group{
				Group: groupResource,
			},
		}

		res = append(res, resource)
	}

	return res, nil
}

func (collector *GCPCollector) ListUsersInGroup(ctx context.Context, group string) ([]string, error) {
	return collector.api.ListUsersInGroup(ctx, group)
}
