package main

import (
	"context"
	"errors"
	"github.com/wtks/cmkp/backend/model"
	"time"
)

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) SetCirclePriorities(ctx context.Context, day int, circleIds []int) (*model.UserCirclePriority, error) {
	return model.SetUserCirclePriorities(ctx, getUserId(ctx), day, circleIds)
}

func (r *mutationResolver) ChangeUserEntries(ctx context.Context, userId int, entries []int) (*model.User, error) {
	if model.IsGranted(ctx, getUserRole(ctx), model.RoleAdmin) {
		return model.ChangeUserEntries(ctx, userId, entries)
	}
	return nil, model.ErrForbidden
}

func (r *mutationResolver) ChangePassword(ctx context.Context, oldPassword string, newPassword string) (bool, error) {
	user, err := model.GetUserByID(ctx, getUserId(ctx))
	if err != nil {
		return false, err
	}

	if !user.CheckPassword(oldPassword) {
		return false, errors.New("パスワードが間違っています")
	}

	if err := model.ChangeUserPassword(ctx, user.ID, newPassword); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) SetCirclePriority(ctx context.Context, day int, rank int, circleId *int) (*model.UserCirclePriority, error) {
	return model.SetUserCirclePriority(ctx, getUserId(ctx), day, rank, circleId)
}

func (r *mutationResolver) ChangeUserEntry(ctx context.Context, userId int, day int, entry bool) (*model.User, error) {
	if model.IsGranted(ctx, getUserRole(ctx), model.RoleAdmin) {
		return model.ChangeUserEntry(ctx, userId, day, entry)
	}
	return nil, model.ErrForbidden
}

func (r *mutationResolver) SetDeadline(ctx context.Context, day int, t time.Time) (time.Time, error) {
	if model.IsGranted(ctx, getUserRole(ctx), model.RolePlanner) {
		dl, err := model.SetDeadline(ctx, day, t)
		if err != nil {
			return time.Time{}, err
		}
		return dl.Datetime, nil
	}
	return time.Time{}, model.ErrForbidden
}

func (r *mutationResolver) CreateItem(ctx context.Context, circleId int, name string, price int) (*model.Item, error) {
	return model.CreateItem(ctx, circleId, name, price)
}

func (r *mutationResolver) ChangeItemName(ctx context.Context, itemId int, name string) (*model.Item, error) {
	if model.IsGranted(ctx, getUserRole(ctx), model.RolePlanner) {
		return model.ChangeItemName(ctx, itemId, name)
	}
	return nil, model.ErrForbidden
}

func (r *mutationResolver) ChangeItemPrice(ctx context.Context, itemId int, price int) (*model.Item, error) {
	return model.ChangeItemPrice(ctx, itemId, price)
}

func (r *mutationResolver) CreateUser(ctx context.Context, username string, displayName string, password string) (*model.User, error) {
	if model.IsGranted(ctx, getUserRole(ctx), model.RoleAdmin) {
		return model.CreateUser(ctx, username, displayName, password)
	}
	return nil, model.ErrForbidden
}

func (r *mutationResolver) ChangeUserPassword(ctx context.Context, userId int, password string) (bool, error) {
	if model.IsGranted(ctx, getUserRole(ctx), model.RoleAdmin) {
		user, err := model.GetUserByID(ctx, userId)
		if err != nil {
			return false, err
		}

		if err := model.ChangeUserPassword(ctx, user.ID, password); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, model.ErrForbidden
}

func (r *mutationResolver) ChangeUserRole(ctx context.Context, userId int, role model.Role) (*model.User, error) {
	if model.IsGranted(ctx, getUserRole(ctx), model.RoleAdmin) {
		return model.ChangeUserRole(ctx, userId, role)
	}
	return nil, model.ErrForbidden
}

func (r *mutationResolver) CreateRequest(ctx context.Context, userId *int, itemId int, num int) (*model.UserRequestItem, error) {
	item, err := model.GetItemByID(ctx, itemId)
	if err != nil {
		return nil, err
	}

	if userId != nil {
		if getUserId(ctx) != *userId && !model.IsGranted(ctx, getUserRole(ctx), model.RolePlanner) {
			return nil, model.ErrForbidden
		}
		if _, err := model.GetUserByID(ctx, *userId); err != nil {
			return nil, err
		}
		return model.CreateUserRequestItem(ctx, *userId, itemId, num)
	} else {
		circle, err := item.Circle(ctx)
		if err != nil {
			return nil, err
		}
		dl, err := model.GetDeadline(ctx, circle.Day)
		if err != nil {
			return nil, err
		}
		if dl.IsOver() {
			return nil, errors.New("deadline is over")
		}

		return model.CreateUserRequestItem(ctx, getUserId(ctx), itemId, num)
	}
}

func (r *mutationResolver) ChangeRequestNum(ctx context.Context, requestId int, num int) (*model.UserRequestItem, error) {
	ur, err := model.GetUserRequestItemByID(ctx, requestId)
	if err != nil {
		return nil, err
	}

	if ur.UserID != getUserId(ctx) && !model.IsGranted(ctx, getUserRole(ctx), model.RolePlanner) {
		return nil, model.ErrForbidden
	}

	circle, err := ur.Circle(ctx)
	if err != nil {
		return nil, err
	}
	dl, err := model.GetDeadline(ctx, circle.Day)
	if err != nil {
		return nil, err
	}
	if dl.IsOver() {
		return nil, errors.New("deadline is over")
	}

	return model.ChangeUserRequestItemNum(ctx, requestId, num)
}

func (r *mutationResolver) DeleteRequest(ctx context.Context, id int) (bool, error) {
	ur, err := model.GetUserRequestItemByID(ctx, id)
	if err != nil {
		return false, err
	}

	if ur.UserID != getUserId(ctx) && !model.IsGranted(ctx, getUserRole(ctx), model.RolePlanner) {
		return false, model.ErrForbidden
	}

	if err := model.DeleteUserRequestItem(ctx, id); err != nil {
		return false, err
	}
	return true, err
}

func (r *mutationResolver) PostRequestNote(ctx context.Context, content string) (*model.UserRequestNote, error) {
	return model.CreateUserRequestNote(ctx, getUserId(ctx), content)
}

func (r *mutationResolver) EditRequestNote(ctx context.Context, id int, content string) (*model.UserRequestNote, error) {
	panic("not implemented")
}

func (r *mutationResolver) DeleteRequestNote(ctx context.Context, id int) (bool, error) {
	note, err := model.GetUserRequestNoteByID(ctx, id)
	if err != nil {
		return false, err
	}
	if note.UserID != getUserId(ctx) {
		return false, errors.New("forbidden")
	}

	if err := model.DeleteUserRequestNote(ctx, id); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) PostCircleMemo(ctx context.Context, circleId int, content string) (*model.CircleMemo, error) {
	return model.CreateCircleMemo(ctx, circleId, getUserId(ctx), content)
}

func (r *mutationResolver) EditCircleMemo(ctx context.Context, id int, content string) (*model.CircleMemo, error) {
	panic("not implemented")
}

func (r *mutationResolver) DeleteCircleMemo(ctx context.Context, id int) (bool, error) {
	memo, err := model.GetCircleMemoByID(ctx, id)
	if err != nil {
		return false, err
	}
	if memo.UserID != getUserId(ctx) {
		return false, errors.New("forbidden")
	}

	if err := model.DeleteCircleMemo(ctx, id); err != nil {
		return false, err
	}
	return true, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) UserRequestedCircles(ctx context.Context, userId int) ([]*model.Circle, error) {
	if userId == getUserId(ctx) || model.IsGranted(ctx, getUserRole(ctx), model.RolePlanner) {
		return model.GetRequestedCirclesByUser(ctx, userId)
	}
	return nil, model.ErrForbidden
}

func (r *queryResolver) CirclePriority(ctx context.Context, userId int, day int) (*model.UserCirclePriority, error) {
	if userId == getUserId(ctx) || model.IsGranted(ctx, getUserRole(ctx), model.RolePlanner) {
		return model.GetUserCirclePriorityByUserIDAndDay(ctx, userId, day)
	}
	return nil, model.ErrForbidden
}

func (r *queryResolver) MyCirclePriorityIds(ctx context.Context, day int) ([]int, error) {
	up, err := model.GetUserCirclePriorityByUserIDAndDay(ctx, getUserId(ctx), day)
	if err != nil {
		return make([]int, 0), nil
	}

	res := make([]int, 0)
	for _, v := range up.Priorities() {
		if v != nil {
			res = append(res, *v)
		}
	}
	return res, nil
}

func (r *queryResolver) Circles(ctx context.Context, q string, days []int) ([]*model.Circle, error) {
	return model.SearchCircles(ctx, q, days)
}

func (r *queryResolver) Deadline(ctx context.Context, day int) (time.Time, error) {
	deadline, err := model.GetDeadline(ctx, day)
	if err != nil {
		return time.Time{}, err
	}
	return deadline.Datetime, nil
}

func (r *queryResolver) Deadlines(ctx context.Context) ([]*model.Deadline, error) {
	return model.GetDeadlines(ctx)
}

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	return model.GetUserByID(ctx, getUserId(ctx))
}

func (r *queryResolver) MyRequests(ctx context.Context) ([]*model.UserRequestItem, error) {
	return model.GetUserRequestItemsByUserID(ctx, getUserId(ctx))
}

func (r *queryResolver) MyRequestNotes(ctx context.Context) ([]*model.UserRequestNote, error) {
	return model.GetUserRequestNotesByUserID(ctx, getUserId(ctx))
}

func (r *queryResolver) MyRequestedCircles(ctx context.Context) ([]*model.Circle, error) {
	return model.GetRequestedCirclesByUser(ctx, getUserId(ctx))
}

func (r *queryResolver) User(ctx context.Context, id int) (*model.User, error) {
	return model.GetUserByID(ctx, id)
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	return model.GetUsers(ctx)
}

func (r *queryResolver) Circle(ctx context.Context, id int) (*model.Circle, error) {
	return model.GetCircleByID(ctx, id)
}

func (r *queryResolver) RequestedCircles(ctx context.Context, day int) ([]*model.Circle, error) {
	if model.IsGranted(ctx, getUserRole(ctx), model.RolePlanner) {
		return model.GetRequestedCirclesByDay(ctx, day)
	}
	return nil, model.ErrForbidden
}

func (r *queryResolver) CircleMemo(ctx context.Context, id int) (*model.CircleMemo, error) {
	return model.GetCircleMemoByID(ctx, id)
}

func (r *queryResolver) CircleMemos(ctx context.Context, circleId int) ([]*model.CircleMemo, error) {
	return model.GetCircleMemosByCircleID(ctx, circleId)
}

func (r *queryResolver) Item(ctx context.Context, id int) (*model.Item, error) {
	return model.GetItemByID(ctx, id)
}

func (r *queryResolver) Items(ctx context.Context, circleId int) ([]*model.Item, error) {
	return model.GetItemsByCircleID(ctx, circleId)
}

func (r *queryResolver) Request(ctx context.Context, id int) (*model.UserRequestItem, error) {
	uri, err := model.GetUserRequestItemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if uri.UserID == getUserId(ctx) || model.IsGranted(ctx, getUserRole(ctx), model.RolePlanner) {
		return uri, nil
	}
	return nil, model.ErrForbidden
}

func (r *queryResolver) RequestNote(ctx context.Context, id int) (*model.UserRequestNote, error) {
	note, err := model.GetUserRequestNoteByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if note.UserID == getUserId(ctx) || model.IsGranted(ctx, getUserRole(ctx), model.RolePlanner) {
		return note, nil
	}
	return nil, model.ErrForbidden
}

func (r *queryResolver) RequestNotes(ctx context.Context, userId int) ([]*model.UserRequestNote, error) {
	if userId > 0 {
		if userId == getUserId(ctx) || model.IsGranted(ctx, getUserRole(ctx), model.RolePlanner) {
			return model.GetUserRequestNotesByUserID(ctx, userId)
		}
		return nil, model.ErrForbidden
	}
	if model.IsGranted(ctx, getUserRole(ctx), model.RolePlanner) {
		return model.GetUserRequestNotes(ctx)
	}
	return nil, model.ErrForbidden
}
