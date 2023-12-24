package announcements

import (
	"context"
	"net"
	"sync"
	"time"

	"p2p-snake/internal/log"
	"p2p-snake/internal/p2p/dto"
)

const (
	announceTimeout = time.Second
)

type Announcement struct {
	lastUpdate time.Time
	addr       *net.UDPAddr

	gameName   string
	width      int32
	height     int32
	foodStatic int32
	stateDelay int32
}

func NewAnnouncement(addr *net.UDPAddr, gameName string, width int32, height int32, foodStatic int32, stateDelay int32) *Announcement {
	return &Announcement{
		lastUpdate: time.Now(),
		addr:       addr,

		gameName:   gameName,
		width:      width,
		height:     height,
		foodStatic: foodStatic,
		stateDelay: stateDelay,
	}
}

func (a Announcement) Addr() *net.UDPAddr {
	return a.addr
}

func (a Announcement) GameName() string {
	return a.gameName
}

func (a Announcement) Width() int32 {
	return a.width
}

func (a Announcement) Height() int32 {
	return a.height
}

func (a Announcement) FoodStatic() int32 {
	return a.foodStatic
}

func (a Announcement) StateDelay() int32 {
	return a.stateDelay
}

type AnnouncementCollector struct {
	announcements map[string]*Announcement

	// Closing
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func NewAnnouncementCollector() *AnnouncementCollector {
	return &AnnouncementCollector{
		announcements: make(map[string]*Announcement),

		cancel: func() {},
		wg:     &sync.WaitGroup{},
	}
}

func (collector *AnnouncementCollector) Start() {
	collector.cancel()
	collector.wg.Wait()

	var ctx context.Context
	ctx, collector.cancel = context.WithCancel(context.Background())
	collector.wg.Add(1)
	go collector.removeExpiredAnnouncement(ctx)
}

func (collector *AnnouncementCollector) AddAnnouncement(announcement *Announcement) {
	if announcedGame, ok := collector.announcements[announcement.gameName]; ok {
		announcedGame.lastUpdate = time.Now()
		log.Logger.Debugf("Announcement \"%s\" updated", announcement.gameName)
	} else {
		collector.announcements[announcement.gameName] = announcement
		log.Logger.Debugf("Announcement \"%s\" added", announcement.gameName)
	}
}

func (collector *AnnouncementCollector) removeExpiredAnnouncement(ctx context.Context) {
	defer collector.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for key, announcement := range collector.announcements {
				if time.Since(announcement.lastUpdate) > announceTimeout {
					delete(collector.announcements, key)
					log.Logger.Debugf("Announcement \"%s\" deleted", announcement.gameName)
				}
			}
			time.Sleep(announceTimeout)
		}
	}
}

func (collector *AnnouncementCollector) GetGameInfoDtos() []dto.GameInfoDto {
	gameInfoDtos := make([]dto.GameInfoDto, 0)
	for _, announcement := range collector.announcements {
		gameInfoDtos = append(gameInfoDtos, dto.NewGameInfoDto(
			announcement.GameName(),
			announcement.Width(),
			announcement.Height(),
			announcement.StateDelay(),
		))
	}
	return gameInfoDtos
}

func (collector *AnnouncementCollector) ExistsAnnouncementByGameName(gameName string) bool {
	_, ok := collector.announcements[gameName]
	return ok
}

func (collector *AnnouncementCollector) FindByGameName(gameName string) (*Announcement, bool) {
	announcement, ok := collector.announcements[gameName]
	if !ok {
		return nil, false
	}
	return announcement, true
}

func (collector *AnnouncementCollector) Close() {
	collector.cancel()
	collector.wg.Wait()
}
