package service

import (
    "testing"

    "github.com/taskirx/go-bidding-engine/internal/model"
)

func BenchmarkCreativeOptimization_SelectCreative(b *testing.B) {
    svc := NewCreativeOptimizationService(nil)
    campaign := &model.Campaign{ID: "campaign-001", BidPrice: 2.0}
    req := &model.BidRequest{ID: "req-001", PublisherID: "pub-001"}
    for i := 0; i < 100; i++ {
        svc.RecordImpression("creative-"+string(rune(i%3)), "")
        if i%10 == 0 { svc.RecordClick("creative-"+string(rune(i%3)), "") }
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ { svc.SelectCreative(campaign, req) }
}

func BenchmarkBidLandscape_AnalyzeLandscape(b *testing.B) {
    svc := NewBidLandscapeService(nil)
    for i := 0; i < 1000; i++ {
        req := &model.BidRequest{PublisherID: "pub-001", Device: model.InternalDevice{Type: "mobile"}}
        svc.RecordBid(req, 1.0+float64(i%50)*0.1, 0.9*(1.0+float64(i%50)*0.1), i%3 == 0)
    }
    campaign := &model.Campaign{ID: "campaign-001", BidPrice: 2.0}
    req := &model.BidRequest{PublisherID: "pub-001", Device: model.InternalDevice{Type: "mobile"}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ { svc.AnalyzeLandscape(campaign, req) }
}

func BenchmarkDynamicBid_CalculateDynamicBid(b *testing.B) {
    svc := NewDynamicBidService(nil)
    campaign := &model.Campaign{ID: "campaign-001", BidPrice: 2.50}
    req := &model.BidRequest{ID: "req-001", PublisherID: "pub-001", Device: model.InternalDevice{Type: "mobile"}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ { svc.CalculateDynamicBid(campaign, req) }
}

func BenchmarkDayparting_CalculateDaypartMultiplier(b *testing.B) {
    svc := NewDaypartingService(nil)
    campaign := &model.Campaign{ID: "campaign-001", BidPrice: 2.0}
    req := &model.BidRequest{ID: "req-001", PublisherID: "pub-001"}
    b.ResetTimer()
    for i := 0; i < b.N; i++ { svc.CalculateDaypartMultiplier(campaign, req) }
}

func BenchmarkBidCache_Get(b *testing.B) {
    svc := NewBidCacheService(nil)
    bid := &CachedBid{Price: 2.50}
    svc.Set(nil, "test-key", bid)
    b.ResetTimer()
    for i := 0; i < b.N; i++ { svc.Get(nil, "test-key") }
}

func BenchmarkChurnPrediction_PredictChurn(b *testing.B) {
    svc := NewChurnPredictionService(nil)
    for i := 0; i < 100; i++ {
        svc.RecordUserActivity("user-001", "purchase", nil)
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ { svc.PredictChurn("user-001") }
}

func BenchmarkPG_GetDeal(b *testing.B) {
    svc := NewProgrammaticGuaranteedService(nil)
    deal := &PGDeal{Name: "Test", BuyerID: "buyer-001", SellerID: "seller-001", CommittedImpressions: 1000000, FixedPrice: 5.00}
    created, _ := svc.CreateDeal(deal)
    for i := 0; i < 1000; i++ { svc.RecordImpression(created.ID, 5.0) }
    b.ResetTimer()
    for i := 0; i < b.N; i++ { svc.GetDeal(created.ID) }
}

func BenchmarkDirectPublisher_GetPublisher(b *testing.B) {
    svc := NewDirectPublisherService(nil)
    pub := &DirectPublisher{Name: "Test", Domain: "example.com"}
    registered, _ := svc.RegisterPublisher(pub)
    b.ResetTimer()
    for i := 0; i < b.N; i++ { svc.GetPublisher(registered.ID) }
}

func BenchmarkCampaign_IsMatch(b *testing.B) {
    campaign := &model.Campaign{ID: "campaign-001", BidPrice: 2.50, Creative: model.Creative{Type: "banner"}, Targeting: model.Targeting{Countries: []string{"US","CA","GB"}, Devices: []string{"mobile","desktop"}, OS: []string{"ios","android"}, MinAge: 18, MaxAge: 65, Categories: []string{"sports","tech"}}}
    req := &model.BidRequest{User: model.InternalUser{Country: "US", Age: 30, Categories: []string{"tech"}}, Device: model.InternalDevice{Type: "mobile", OS: "ios"}, AdSlot: model.AdSlot{Formats: []string{"banner"}}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ { campaign.IsMatch(req) }
}

func BenchmarkCampaign_IsMatch_GeoFence(b *testing.B) {
    campaign := &model.Campaign{ID: "campaign-001", BidPrice: 2.50, Creative: model.Creative{Type: "banner"}, Targeting: model.Targeting{GeoFences: []model.GeoFence{{Lat: 40.7128, Lon: -74.0060, Radius: 10.0, Name: "NYC"}}}}
    req := &model.BidRequest{Device: model.InternalDevice{Type: "mobile", Geo: model.InternalGeo{Lat: 40.7200, Lon: -74.0100}}, AdSlot: model.AdSlot{Formats: []string{"banner"}}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ { campaign.IsMatch(req) }
}

func BenchmarkBidLandscape_Parallel(b *testing.B) {
    svc := NewBidLandscapeService(nil)
    for i := 0; i < 1000; i++ {
        req := &model.BidRequest{PublisherID: "pub-001", Device: model.InternalDevice{Type: "mobile"}}
        svc.RecordBid(req, 1.0+float64(i%50)*0.1, 0.9*(1.0+float64(i%50)*0.1), i%3 == 0)
    }
    campaign := &model.Campaign{ID: "campaign-001", BidPrice: 2.0}
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        req := &model.BidRequest{PublisherID: "pub-001", Device: model.InternalDevice{Type: "mobile"}}
        for pb.Next() { svc.AnalyzeLandscape(campaign, req) }
    })
}

func BenchmarkDynamicBid_Parallel(b *testing.B) {
    svc := NewDynamicBidService(nil)
    campaign := &model.Campaign{ID: "campaign-001", BidPrice: 2.50}
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        req := &model.BidRequest{ID: "req-001", PublisherID: "pub-001", Device: model.InternalDevice{Type: "mobile"}}
        for pb.Next() { svc.CalculateDynamicBid(campaign, req) }
    })
}
