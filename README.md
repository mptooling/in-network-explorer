# Project: In-Network-Explorer

Sub-title: AI-Powered LinkedIn Prospecting Engine (Go + AWS + LLM)

 1. [Quickstart](docs/quickstart.md) — get running locally in under five minutes.

 2. Project Overview
In-Network-Explorer is a specialized lead-generation and networking tool for IT professionals in Berlin. It uses a Go-based scraper to identify relevant local peers, evaluates their professional alignment using AI, and drafts hyper-personalized connection requests.

The Golden Rule: The system prepares the data; the human performs the final outreach to maintain account integrity.

2. Technical Stack
Language: Go 1.25+

Browser Automation: go-rod/rod (Headless Chromium)

Infrastructure: AWS (Lambda via Docker/ECR, DynamoDB, Secrets Manager)

AI Engine: Amazon Bedrock (Claude 3.5 Sonnet) or OpenAI API

Scheduling: AWS EventBridge

3. Sophisticated "Under the Radar" Logic
To avoid detection, the tool must mimic human browsing patterns:

Multi-Step Warming: Never message on first sight.

Day 1: Visit Profile (Trigger "Someone viewed your profile" notification).

Day 2: Like a recent post or "wait" state.

Day 3: Propose the message to the User.

Randomized Jitter: All browser actions (scrolling, clicking, waiting) must include a random time.Sleep with a variance of 20-40%.

Residential IP Simulation: (Optional/Future) Support for proxy rotation if running outside of local Berlin home IP.

4. Functional Requirements
Phase A: Scraper & Parser (Go + Rod)
Login via session cookies (stored in AWS Secrets Manager) to avoid MFA triggers.

Monitor specific "Berlin IT" keywords and hashtags (e.g., #BerlinTech, #Golang, #BerlinStartups).

Extract: Profile URL, Name, Current Role, Location, and the text of their 3 most recent posts.

Phase B: AI Analysis Engine
Worthiness Score: The LLM must analyze extracted data against a "Target Persona" (e.g., "Senior Software Engineer in Berlin").

Scoring Criteria: * Is the user based in Berlin/Brandenburg?

Are they in a relevant field (DevOps, Go, Cloud)?

Have they posted recently (active user)?

Message Drafting: Generate a 150-300 character message.

Instruction: "Reference a specific detail from their recent post. Use a casual, professional tone. Avoid 'corporate speak'. Mention being a fellow Go dev in Berlin."

Phase C: Persistence (DynamoDB)
Schema:

PK: LinkedIn Profile URL (String)

Status: (Scanned | Liked | Drafted | Sent)

WorthinessScore: (1-10)

DraftedMessage: (String)

LastActionDate: (Timestamp)

Phase D: Human-in-the-Loop Interface
Every 48 hours, the system generates a "Prospect Report" (JSON or simple HTML page).

The report provides: Profile Link + The Drafted Message + "Copy to Clipboard" button.

5. Deployment Specs (AWS)
Containerization: Dockerfile to bundle Go binary + Chromium + dependencies.

Lambda Configuration: 2048MB RAM, 1-minute timeout.

Security: IAM roles for DynamoDB and Bedrock access; credentials never stored in code.

6. Success Metrics
Identify 10 high-quality leads per week.

Zero "Proof of Life" (CAPTCHA) challenges from LinkedIn.

Draft messages that feel indistinguishable from manual research.

Suggested Execution Plan for Claude/AI
Step 1: Write the Go struct definitions for the DynamoDB schema.

Step 2: Build the rod scraper logic for "Profile Visiting" and "Post Extraction."

Step 3: Implement the LLM prompt engineering for the "Worthiness Score."

Step 4: Set up the Docker/AWS Lambda deployment script.
