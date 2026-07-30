package service

import (
	"context"
	"log"
)

// ZNSNotifier sends Zalo Notification Service messages.
// MVP uses a stub that only logs; swap for a real ZNS client later.
type ZNSNotifier interface {
	NotifyMemberRegistered(ctx context.Context, tenantCode string, userID int64, phone, fullName string)
	NotifyCourseRegistered(ctx context.Context, tenantCode string, userID int64, phone, courseTitle string)
}

type stubZNSNotifier struct{}

func NewStubZNSNotifier() ZNSNotifier {
	return &stubZNSNotifier{}
}

func (s *stubZNSNotifier) NotifyMemberRegistered(_ context.Context, tenantCode string, userID int64, phone, fullName string) {
	log.Printf("[zns-stub] member_registered tenant=%s user_id=%d phone=%s name=%s", tenantCode, userID, phone, fullName)
}

func (s *stubZNSNotifier) NotifyCourseRegistered(_ context.Context, tenantCode string, userID int64, phone, courseTitle string) {
	log.Printf("[zns-stub] course_registered tenant=%s user_id=%d phone=%s course=%s", tenantCode, userID, phone, courseTitle)
}
