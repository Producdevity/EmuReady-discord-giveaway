package handlers

import "github.com/gofiber/fiber/v2"

func NewTermsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Type("html").SendString(termsOfServiceHTML)
	}
}

const termsOfServiceHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="robots" content="index, follow">
  <title>Terms of Service | EmuReady Discord Giveaway</title>
  <meta name="description" content="Terms of Service for the EmuReady Discord Giveaway bot.">
  <style>
    :root {
      color-scheme: light dark;
      --background: #ffffff;
      --foreground: #111827;
      --muted: #4b5563;
      --text: #374151;
      --border: #e5e7eb;
      --link: #2563eb;
    }

    @media (prefers-color-scheme: dark) {
      :root {
        --background: #111217;
        --foreground: #f9fafb;
        --muted: #9ca3af;
        --text: #d1d5db;
        --border: #30323d;
        --link: #8ab4ff;
      }
    }

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      background: var(--background);
      color: var(--foreground);
      font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 18px;
      line-height: 1.65;
    }

    main {
      width: min(896px, 100%);
      margin: 0 auto;
      padding: 32px 16px 56px;
    }

    h1 {
      margin: 0 0 32px;
      color: var(--foreground);
      font-size: 40px;
      font-weight: 800;
      line-height: 1.1;
    }

    h2 {
      margin: 0 0 16px;
      color: var(--foreground);
      font-size: 24px;
      font-weight: 700;
      line-height: 1.3;
    }

    p {
      margin: 0 0 16px;
      color: var(--text);
    }

    .updated {
      margin-bottom: 32px;
      color: var(--muted);
    }

    section {
      margin-top: 32px;
    }

    ul {
      margin: 0;
      padding-left: 28px;
      color: var(--text);
    }

    li + li {
      margin-top: 8px;
    }

    a {
      color: var(--link);
      text-underline-offset: 3px;
    }

    strong {
      color: var(--foreground);
      font-weight: 700;
    }

    code {
      padding: 2px 6px;
      border: 1px solid var(--border);
      border-radius: 6px;
      font-size: 0.9em;
    }
  </style>
</head>
<body>
  <main>
    <h1>Terms of Service</h1>

    <p class="updated"><strong>Last updated:</strong> June 13, 2026</p>

    <section>
      <h2>1. Acceptance of Terms</h2>
      <p>
        By using the EmuReady Discord Giveaway bot, including commands such as
        <code>/enter-giveaway</code>, <code>/entrants</code>, and <code>/winner</code>, you agree
        to these Terms of Service. If you do not agree, do not use the giveaway bot.
      </p>
    </section>

    <section>
      <h2>2. Description of Service</h2>
      <p>
        The EmuReady Discord Giveaway bot helps run giveaways in the EmuReady Discord server. The
        bot can verify that a participant has starred the configured GitHub repository, assign a
        giveaway role, count entrants, and help staff draw winners from current eligible entrants.
      </p>
      <p>
        The bot is provided for EmuReady community giveaways. It is not a general-purpose
        sweepstakes platform.
      </p>
    </section>

    <section>
      <h2>3. Eligibility</h2>
      <p>To enter a giveaway through the bot, you must:</p>
      <ul>
        <li>Be a member of the relevant Discord server.</li>
        <li>Comply with Discord&apos;s Terms of Service and Community Guidelines.</li>
        <li>Comply with GitHub&apos;s Terms of Service when using GitHub OAuth or starring a repository.</li>
        <li>Meet any eligibility rules stated in the applicable giveaway announcement.</li>
        <li>Be legally allowed to participate where you live. Giveaways are void where prohibited.</li>
      </ul>
    </section>

    <section>
      <h2>4. Giveaway Entry and Fair Use</h2>
      <p>
        Entry may require using <code>/enter-giveaway</code>, authorizing with GitHub, and keeping
        the configured GitHub repository starred until winner selection. Unless a giveaway
        announcement says otherwise, one GitHub account may be linked to one Discord account for
        giveaway entry.
      </p>
      <p>You agree not to:</p>
      <ul>
        <li>Use alternate accounts, fake accounts, automation, or other manipulation to gain extra entries.</li>
        <li>Interfere with the bot, the giveaway process, or the underlying Discord or GitHub services.</li>
        <li>Attempt to bypass eligibility checks, role checks, or duplicate-entry protections.</li>
        <li>Misrepresent your identity, account ownership, or eligibility.</li>
      </ul>
      <p>
        EmuReady staff may remove entries, remove giveaway roles, redraw winners, or disqualify
        participants who violate these terms, server rules, giveaway rules, or platform policies.
      </p>
    </section>

    <section>
      <h2>5. Winner Selection</h2>
      <p>
        Winners are selected from the current eligible entrant pool available to the bot at the time
        staff run the winner command. Before or during winner selection, the bot or staff may
        re-check eligibility, including GitHub star status and Discord role membership.
      </p>
      <p>
        A selected winner may be disqualified and replaced if they no longer meet giveaway rules,
        cannot be contacted, decline the prize, or are found to have manipulated the process.
      </p>
    </section>

    <section>
      <h2>6. Third-Party Services</h2>
      <p>
        The bot depends on Discord, GitHub, and Railway. Your use of those services remains subject
        to their own terms and policies. The giveaway bot is not sponsored, endorsed, or administered
        by Discord or GitHub.
      </p>
    </section>

    <section>
      <h2>7. Privacy</h2>
      <p>
        Use of the giveaway bot is also governed by our
        <a href="/privacy">Privacy Policy</a>, which explains what Discord and GitHub data is used
        for giveaway entry, role assignment, entrant counts, winner selection, retention, and
        deletion requests.
      </p>
    </section>

    <section>
      <h2>8. Availability and Changes</h2>
      <p>
        The bot is provided as-is. We may update, suspend, or discontinue the bot, change giveaway
        rules, fix errors, or modify these terms when needed for security, reliability, platform
        requirements, or community operations.
      </p>
    </section>

    <section>
      <h2>9. Limitation of Liability</h2>
      <p>
        To the fullest extent permitted by law, EmuReady is not liable for indirect, incidental,
        special, consequential, or punitive damages related to use of the giveaway bot, giveaway
        participation, third-party service outages, missed entries, or unavailable commands.
      </p>
    </section>

    <section>
      <h2>10. Contact</h2>
      <p>
        For questions about these Terms of Service or a giveaway run through the bot, contact
        EmuReady staff in the EmuReady Discord server.
      </p>
    </section>
  </main>
</body>
</html>`
