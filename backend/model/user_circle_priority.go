package model

import (
	"context"
	"time"
)

type UserCirclePriority struct {
	ID        int `gorm:"primary_key"`
	UserID    int `gorm:"unique_index:user_day"`
	Day       int `gorm:"unique_index:user_day"`
	Priority1 *int
	Priority2 *int
	Priority3 *int
	Priority4 *int
	Priority5 *int
	CreatedAt time.Time `gorm:"precision:6"`
	UpdatedAt time.Time `gorm:"precision:6"`
}

func (ucp *UserCirclePriority) Priority(rank int) *int {
	switch rank {
	case 1:
		return ucp.Priority1
	case 2:
		return ucp.Priority2
	case 3:
		return ucp.Priority3
	case 4:
		return ucp.Priority4
	case 5:
		return ucp.Priority5
	default:
		return nil
	}
}

func (ucp *UserCirclePriority) Priorities() []*int {
	return []*int{
		ucp.Priority1,
		ucp.Priority2,
		ucp.Priority3,
		ucp.Priority4,
		ucp.Priority5,
	}
}

func (ucp *UserCirclePriority) User(ctx context.Context) (*User, error) {
	return GetUserByID(ctx, ucp.UserID)
}

func (ucp *UserCirclePriority) getPriorityRank(rank int) *PriorityRank {
	id := ucp.Priority(rank)
	if id == nil {
		return nil
	}
	return &PriorityRank{
		CircleID: *id,
		UserID:   ucp.UserID,
		Rank:     rank,
	}
}

type PriorityRank struct {
	CircleID int
	UserID   int
	Rank     int
}

func (pr *PriorityRank) Circle(ctx context.Context) (*Circle, error) {
	return GetCircleByID(ctx, pr.CircleID)
}

func (pr *PriorityRank) User(ctx context.Context) (*User, error) {
	return GetUserByID(ctx, pr.UserID)
}

func GetUserCirclePriorityByUserID(ctx context.Context, id int) ([]*UserCirclePriority, error) {
	ucp := make([]*UserCirclePriority, 0)
	if err := orm(ctx).Where(&UserCirclePriority{UserID: id}).Find(ucp).Error; err != nil {
		panic(err)
	}
	return ucp, nil
}

func GetPriorityRankListByCircleID(ctx context.Context, id int) ([]*PriorityRank, error) {
	ucp := make([]*UserCirclePriority, 0)
	if err := orm(ctx).Where("priority1 = ? OR priority2 = ? OR priority3 = ? OR priority4 = ? OR priority5 = ?", id, id, id, id, id).Find(ucp).Error; err != nil {
		panic(err)
	}

	result := make([]*PriorityRank, 0)
	for _, v := range ucp {
		rank := 0
		switch {
		case v.Priority1 != nil && id == *v.Priority1:
			rank = 1
		case v.Priority2 != nil && id == *v.Priority2:
			rank = 2
		case v.Priority3 != nil && id == *v.Priority3:
			rank = 3
		case v.Priority4 != nil && id == *v.Priority4:
			rank = 4
		case v.Priority5 != nil && id == *v.Priority5:
			rank = 5
		}
		result = append(result, v.getPriorityRank(rank))
	}
	return result, nil
}

func SetUserCirclePriority(ctx context.Context, userID, day, rank int, circleID *int) (*UserCirclePriority, error) {
	db := orm(ctx)
	ucp := &UserCirclePriority{}
	if err := db.Where(&UserCirclePriority{UserID: userID, Day: day}).FirstOrCreate(ucp).Error; err != nil {
		panic(err)
	}

	switch rank {
	case 1:
		ucp.Priority1 = circleID
	case 2:
		ucp.Priority2 = circleID
	case 3:
		ucp.Priority3 = circleID
	case 4:
		ucp.Priority4 = circleID
	case 5:
		ucp.Priority5 = circleID
	}

	if err := db.Save(ucp).Error; err != nil {
		return nil, err
	}
	return ucp, nil
}
