// Package linkedin implements the BrowserClient interface using Rod for
// headless browser automation against LinkedIn. All CSS selectors and URL
// patterns are centralised here for easy updating when LinkedIn changes markup.
package linkedin

import "time"

// ── URL patterns ────────────────────────────────────────────────────────────

const (
	baseURL    = "https://www.linkedin.com"
	searchPath = "/search/results/people/"
	loginPath  = "/login"
	feedPath   = "/feed"
)

// berlinGeoUrn is LinkedIn's internal geo identifier for Berlin.
const berlinGeoUrn = "103035651"

// ── Profile page selectors ──────────────────────────────────────────────────

// extractProfileJS is evaluated on a profile page. It returns an object with
// the extracted fields. Selectors target LinkedIn's current React-rendered DOM.
const extractProfileJS = `() => {
	const txt = (sel) => {
		const el = document.querySelector(sel);
		return el ? el.textContent.trim() : "";
	};

	const topCard = document.querySelector('.pv-top-card');
	const name = topCard
		? (topCard.querySelector('h1') || {}).textContent || ""
		: "";

	const headline = txt('.text-body-medium.break-words');
	const location = txt('.pv-text-details__left-panel .text-body-small.inline');

	const aboutSection = document.querySelector('#about');
	let about = "";
	if (aboutSection) {
		const container = aboutSection.closest('section');
		if (container) {
			const span = container.querySelector('.pv-shared-text-with-see-more span[aria-hidden="true"]')
				|| container.querySelector('.inline-show-more-text span[aria-hidden="true"]');
			about = span ? span.textContent.trim() : "";
		}
	}

	const postEls = document.querySelectorAll('.profile-creator-shared-feed-update__mini-container a[href*="/feed/update/"]');
	const posts = [];
	postEls.forEach(el => {
		const text = el.closest('.feed-mini-update')?.textContent?.trim() || el.textContent.trim();
		if (text) posts.push(text.substring(0, 500));
	});

	return {
		name: name.trim(),
		headline: headline,
		location: location,
		about: about,
		posts: posts.slice(0, 5)
	};
}`

// ── Search results selectors ────────────────────────────────────────────────

const extractSearchResultsJS = `() => {
	const links = document.querySelectorAll('.entity-result__title-text a.app-aware-link[href*="/in/"]');
	const urls = [];
	links.forEach(a => {
		const href = a.href.split('?')[0];
		if (href && !urls.includes(href)) urls.push(href);
	});
	return urls;
}`

// ── Block detection selectors ───────────────────────────────────────────────

const detectBlockJS = `() => {
	const url = window.location.href;
	if (url.includes('/authwall') || url.includes('/login') || url.includes('/uas/login')) {
		return "authwall";
	}
	if (document.querySelector('#captcha-challenge')
		|| document.querySelector('.challenge-dialog')
		|| url.includes('/checkpoint/challenge')) {
		return "challenge";
	}
	const main = document.querySelector('main') || document.querySelector('#main');
	if (main && main.textContent.trim().length < 50) {
		return "soft_empty";
	}
	return "none";
}`

// ── Like button selectors ───────────────────────────────────────────────────

const findLikeButtonJS = `() => {
	const btn = document.querySelector(
		'button.react-button__trigger[aria-label*="Like"], ' +
		'button[aria-label*="Like"][aria-pressed="false"]'
	);
	return btn ? true : false;
}`

const clickLikeButtonJS = `() => {
	const btn = document.querySelector(
		'button.react-button__trigger[aria-label*="Like"], ' +
		'button[aria-label*="Like"][aria-pressed="false"]'
	);
	if (btn) { btn.click(); return true; }
	return false;
}`

// ── Timing defaults ─────────────────────────────────────────────────────────

const (
	pageLoadTimeout = 30 * time.Second
	navigationDelay = 2 * time.Second
	scrollPixels    = 1500.0
	activitySuffix  = "/recent-activity/all/"
)
