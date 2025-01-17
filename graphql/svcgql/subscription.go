package svcgql

import (
	"context"

	"go.stevenxie.me/api/v2/graphql"
	"go.stevenxie.me/api/v2/music"
	"go.stevenxie.me/api/v2/music/musicgql"
)

func newSubscriptionResolver(strms Streamers) graphql.SubscriptionResolver {
	return subscriptionResolver{
		music: musicgql.NewSubscriptionResolver(strms.Music),
	}
}

type subscriptionResolver struct {
	music musicgql.SubscriptionResolver
}

var _ graphql.SubscriptionResolver = (*subscriptionResolver)(nil)

func (res subscriptionResolver) Music(ctx context.Context) (
	<-chan *music.CurrentlyPlaying, error) {
	return res.music.CurrentlyPlaying(ctx)
}
