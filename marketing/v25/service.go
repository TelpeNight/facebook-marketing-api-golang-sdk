package v25

import (
	"context"
	"errors"

	"github.com/TelpeNight/facebook-marketing-api-golang-sdk/fb"
)

// Version of the graph API being used.
const Version = "v25.0"

// Service interacts with the Facebook Marketing API.
type Service struct {
	*fb.Client
	AdAccounts        *AdAccountService
	AdCreatives       *AdCreativeService
	Adsets            *AdsetService
	Ads               *AdService
	Audiences         *AudienceService
	Campaigns         *CampaignService
	CustomConversions *CustomConversionService
	Events            *EventService
	Insights          *InsightsService
	Interests         *InterestService
	Images            *ImageService
	Pages             *PageService
	Posts             *PostService
	Search            *SearchService
	Videos            *VideoService
}

// New initializes a new Service and all the Services contained.
func New(accessToken string, opts ...fb.ClientOpt) *Service {
	c := fb.NewClient(accessToken, opts...)
	return &Service{
		Client:            c,
		AdAccounts:        &AdAccountService{c},
		AdCreatives:       &AdCreativeService{c, fb.NewStatsContainer()},
		Adsets:            &AdsetService{c},
		Ads:               &AdService{c},
		Audiences:         &AudienceService{c},
		Campaigns:         &CampaignService{c},
		CustomConversions: &CustomConversionService{c},
		Events:            &EventService{c},
		Insights:          newInsightsService(c.GetLogger(), c),
		Interests:         &InterestService{c},
		Images:            &ImageService{c},
		Pages:             &PageService{c},
		Posts:             &PostService{c, fb.NewStatsContainer()},
		Search:            &SearchService{c},
		Videos:            &VideoService{c},
	}
}

func (s *Service) Ping(ctx context.Context) error {
	return s.GetJSON(ctx, fb.NewRoute(Version, "/me").String(), &struct{}{})
}

// GetMetadata returns the metadata of a graph API object.
func (s *Service) GetMetadata(ctx context.Context, id string) (*fb.Metadata, error) {
	res := &fb.MetadataContainer{}
	err := s.Client.GetJSON(ctx, fb.NewRoute(Version, "/%s", id).Metadata(true).String(), res)
	if err != nil {
		return nil, err
	} else if res.Metadata == nil {
		return nil, errors.New("could not get metadata")
	}

	return res.Metadata, nil
}
