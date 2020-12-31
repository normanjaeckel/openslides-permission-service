package collection

import (
	"context"
	"errors"
	"fmt"

	"github.com/OpenSlides/openslides-permission-service/internal/dataprovider"
	"github.com/OpenSlides/openslides-permission-service/internal/perm"
)

// Mediafile implements the permission for the mediafile collection.
func Mediafile(dp dataprovider.DataProvider) perm.ConnecterFunc {
	read := func(ctx context.Context, userID int, fqfields []perm.FQField, result map[string]bool) error {
		return perm.AllFields(fqfields, result, func(fqfield perm.FQField) (bool, error) {
			fqid := fmt.Sprintf("mediafile/%d", fqfield.ID)
			meetingID, err := dp.MeetingFromModel(ctx, fqid)
			if err != nil {
				var errDoesNotExist dataprovider.DoesNotExistError
				if errors.As(err, &errDoesNotExist) {
					return false, nil
				}
				return false, fmt.Errorf("getting meetingID from model %s: %w", fqid, err)
			}

			// TODO: The following code is the same as EnsurePerm but keeps the perms object.
			//////////////////////////////////////////////////
			committeeID, err := dp.CommitteeID(ctx, meetingID)
			if err != nil {
				return false, fmt.Errorf("getting committee id for meeting: %w", err)
			}

			committeeManager, err := dp.IsManager(ctx, userID, committeeID)
			if err != nil {
				return false, fmt.Errorf("check for manager: %w", err)
			}
			if committeeManager {
				return true, nil
			}

			isMeeting, err := dp.InMeeting(ctx, userID, meetingID)
			if err != nil {
				return false, fmt.Errorf("Looking for user %d in meeting %d: %w", userID, meetingID, err)
			}
			if !isMeeting {
				return false, nil
			}

			perms, err := perm.Perms(ctx, userID, meetingID, dp)
			if err != nil {
				return false, fmt.Errorf("getting user permissions: %w", err)
			}

			hasPerms := perms.HasOne("mediafile.can_manage")
			if hasPerms {
				return true, nil
			}
			//////////////////////////////////////////////////

			var isPublic bool
			field := fqid + "/is_public"
			if err := dp.Get(ctx, field, &isPublic); err != nil {
				return false, fmt.Errorf("get %s: %w", field, err)
			}

			if !isPublic {
				return false, nil
			}

			return perms.HasOne("mediafile.can_see"), nil
		})
	}

	return func(s perm.HandlerStore) {
		s.RegisterReadHandler("mediafile", perm.ReadeCheckerFunc(read))
	}
}