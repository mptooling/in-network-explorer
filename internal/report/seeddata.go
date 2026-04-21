package report

import (
	"time"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

// SeedReport returns a ProspectReport with realistic-looking fake data for
// development preview. It covers visual edge cases: high/medium/low scores,
// long text, missing fields, and special characters.
func SeedReport() *explorer.ProspectReport {
	now := time.Now()
	return &explorer.ProspectReport{
		GeneratedAt: now,
		Prospects: []explorer.ReportItem{
			{
				ProfileURL:      "https://linkedin.com/in/lena-kowalski",
				Name:            "Lena Kowalski",
				Headline:        "Senior Platform Engineer at Delivery Hero",
				Location:        "Berlin, Germany",
				WorthinessScore: 9,
				CalibratedProb:  0.87,
				DraftedMessage:  "Hi Lena, your post on migrating to a service mesh really resonated — I went through a similar journey at my last role and would love to compare notes on the Berlin tech scene.",
				RecentPost:      "Just shipped our new observability stack across 40+ microservices. Grafana + OpenTelemetry is the way.",
				ScannedAt:       now.Add(-48 * time.Hour),
			},
			{
				ProfileURL:      "https://linkedin.com/in/markus-braun",
				Name:            "Markus Braun",
				Headline:        "Staff Backend Engineer — Distributed Systems at Zalando SE",
				Location:        "Berlin, Germany",
				WorthinessScore: 8,
				CalibratedProb:  0.72,
				DraftedMessage:  "Hey Markus, your deep-dive on event sourcing at Zalando was eye-opening. I'm building something similar and would value your perspective.",
				RecentPost:      "Event sourcing at scale: lessons learned after 3 years in production",
				ScannedAt:       now.Add(-36 * time.Hour),
			},
			{
				ProfileURL:      "https://linkedin.com/in/sophie-nguyen",
				Name:            "Sophie Nguyen",
				Headline:        "Engineering Manager, Infrastructure & Developer Experience at SoundCloud",
				Location:        "Berlin, Germany",
				WorthinessScore: 7,
				CalibratedProb:  0.55,
				DraftedMessage:  "Hi Sophie, your talk on developer experience metrics at the Berlin DevOps meetup was inspiring — especially the part about reducing CI feedback loops.",
				RecentPost:      "Thrilled to announce we cut our CI pipeline from 25 min to 8 min. Here's how we did it.",
				ScannedAt:       now.Add(-60 * time.Hour),
			},
			{
				ProfileURL:      "https://linkedin.com/in/felix-richter",
				Name:            "Felix Richter",
				Headline:        "DevOps Lead at N26",
				Location:        "Berlin, Germany",
				WorthinessScore: 5,
				CalibratedProb:  0.38,
				DraftedMessage:  "Hey Felix, your GitOps workflow post caught my eye — would love to exchange ideas on Terraform patterns for fintech.",
				RecentPost:      "Terraform tips for financial services: compliance as code",
				ScannedAt:       now.Add(-72 * time.Hour),
			},
			{
				ProfileURL:      "https://linkedin.com/in/anna-mueller",
				Name:            "Anna Müller",
				Headline:        "Cloud Architect at Deutsche Telekom IT Solutions — building next-gen telco infrastructure on AWS & Kubernetes across multiple regions",
				Location:        "Berlin, Brandenburg, Germany",
				WorthinessScore: 6,
				CalibratedProb:  0.44,
				DraftedMessage:  "Hi Anna, I saw your post about multi-region Kubernetes at Deutsche Telekom — impressive scale. I'd love to connect and chat about cloud architecture challenges in Berlin.",
				RecentPost:      "Lessons from running Kubernetes clusters across 5 AWS regions for Deutsche Telekom's B2B platform. Spoiler: DNS is still the hardest part.",
				ScannedAt:       now.Add(-50 * time.Hour),
			},
			{
				ProfileURL:      "https://linkedin.com/in/jorge-santos",
				Name:            "Jorge Santos",
				Headline:        "Junior Software Engineer at Gorillas (acquired by Getir)",
				Location:        "Berlin, Germany",
				WorthinessScore: 3,
				CalibratedProb:  0.10,
				DraftedMessage:  "Hey Jorge, cool that you're working in the quick-commerce space in Berlin. Happy to connect!",
				RecentPost:      "Started my first week at the new office!",
				ScannedAt:       now.Add(-96 * time.Hour),
			},
			{
				ProfileURL:      "https://linkedin.com/in/elena-volkova",
				Name:            "Elena Volkova",
				Headline:        "SRE & Chaos Engineering at Personio",
				Location:        "Berlin, Germany",
				WorthinessScore: 10,
				CalibratedProb:  0.93,
				DraftedMessage:  "Hi Elena, your chaos engineering experiments at Personio are fascinating — particularly the game day format. I'm exploring similar reliability practices and would love to learn from your approach over coffee in Berlin.",
				RecentPost:      "We ran our biggest game day yet: 200 engineers, 3 hours, 12 failure scenarios. Here's what broke (and what didn't).",
				ScannedAt:       now.Add(-24 * time.Hour),
			},
		},
	}
}
