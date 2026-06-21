package handlers

import "github.com/gofiber/fiber/v2"

func NewPrivacyHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Type("html").SendString(privacyPolicyHTML)
	}
}

const privacyPolicyHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="robots" content="index, follow">
  <title>Privacy Policy | EmuReady Discord Giveaway</title>
  <meta name="description" content="Privacy policy for the EmuReady Discord Giveaway bot.">
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

    h3 {
      margin: 24px 0 12px;
      color: var(--foreground);
      font-size: 20px;
      font-weight: 600;
      line-height: 1.35;
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
    <h1>Privacy Policy</h1>

    <p class="updated"><strong>Last updated:</strong> June 13, 2026</p>

    <section>
      <h2>1. Introduction</h2>
      <p>
        This Privacy Policy explains how EmuReady Discord Giveaway ("we", "our", or "us")
        collects, uses, and protects information when you use the EmuReady Discord Giveaway bot.
        This policy applies only to the giveaway bot and its related OAuth callback pages. Other
        EmuReady services may have their own privacy policies.
      </p>
      <p>
        The giveaway bot lets members of the EmuReady Discord server enter a giveaway after
        verifying that their GitHub account has starred the configured EmuReady repository.
      </p>
    </section>

    <section>
      <h2>2. Information We Collect</h2>

      <h3>2.1 Information from Discord</h3>
      <ul>
        <li><strong>Discord user ID:</strong> Used to identify the Discord account entering the giveaway.</li>
        <li><strong>Discord server and role data:</strong> Used to assign the configured giveaway ping role for known entrants.</li>
        <li><strong>Slash command interaction data:</strong> Used only to respond to giveaway commands such as <code>/enter-giveaway</code>, <code>/entrants</code>, and <code>/winner</code>.</li>
      </ul>

      <h3>2.2 Information from GitHub</h3>
      <ul>
        <li><strong>GitHub user ID and username:</strong> Used to link one GitHub account to one Discord account for giveaway eligibility.</li>
        <li><strong>Repository star status:</strong> Used to confirm that your GitHub account has starred the configured repository.</li>
      </ul>

      <h3>2.3 Technical Information</h3>
      <ul>
        <li><strong>OAuth state:</strong> A short-lived signed value used to protect the GitHub authorization flow.</li>
        <li><strong>Operational logs:</strong> Request IDs, timestamps, and error details needed to operate and troubleshoot the service.</li>
      </ul>

      <p>
        We do not collect Discord message content, direct messages, presence data, passwords,
        Discord tokens from users, GitHub passwords, or GitHub private repository contents.
      </p>
    </section>

    <section>
      <h2>3. How We Use Your Information</h2>
      <p>We use the information above to:</p>
      <ul>
        <li>Verify giveaway eligibility through GitHub star status.</li>
        <li>Prevent the same GitHub account from entering through multiple Discord accounts.</li>
        <li>Assign the configured giveaway ping role in Discord.</li>
        <li>Count giveaway entrants and draw winners from stored giveaway entries.</li>
        <li>Re-check eligibility before a winner is selected.</li>
        <li>Operate, secure, debug, and improve the giveaway service.</li>
      </ul>
    </section>

    <section>
      <h2>4. Information Sharing</h2>
      <p>
        We do not sell, rent, trade, or monetize personal information. Information is shared only
        when needed to operate the giveaway:
      </p>
      <ul>
        <li><strong>Discord:</strong> The bot uses Discord APIs to receive commands and assign the giveaway ping role.</li>
        <li><strong>GitHub:</strong> The bot uses GitHub OAuth and GitHub APIs to verify star status.</li>
        <li><strong>Railway:</strong> The service and database are hosted on Railway infrastructure.</li>
        <li><strong>Public Discord output:</strong> Winner results or staff command responses may be visible in the Discord channel where staff run the command.</li>
        <li><strong>Legal or safety requirements:</strong> We may disclose information if required by law or necessary to protect the service, users, or the EmuReady community.</li>
      </ul>
    </section>

    <section>
      <h2>5. Data Retention</h2>
      <p>
        Active giveaway entries are stored while they are needed to run the giveaway, prevent
        duplicate entries, audit winner selection, and handle support requests. When staff reset a
        giveaway, entries are archived with a deletion timestamp and excluded from future entrant
        counts and winner draws. Archived entries may remain until a deletion request is processed
        or staff permanently clean up historical giveaway data.
      </p>
      <p>
        OAuth state data is short-lived and expires automatically. Operational logs are retained
        only as needed for hosting, debugging, abuse prevention, and service reliability.
      </p>
    </section>

    <section>
      <h2>6. Data Security</h2>
      <p>
        We use reasonable technical and organizational safeguards for the giveaway service,
        including HTTPS, environment-variable secrets, restricted access to production systems,
        and a managed PostgreSQL database. We do not store user passwords.
      </p>
      <p>
        No internet-based service can be guaranteed to be completely secure. If we learn of
        unauthorized access that affects giveaway data, we will take appropriate action.
      </p>
    </section>

    <section>
      <h2>7. Your Rights and Choices</h2>
      <p>
        You can choose not to enter the giveaway. If you have already entered, you may ask EmuReady
        staff in the EmuReady Discord server to remove your giveaway entry and related stored data.
      </p>
      <p>
        You can also manage giveaway ping notifications through the Discord server&apos;s role
        settings where available, unstar the GitHub repository, revoke GitHub OAuth access in your
        GitHub account settings, or remove the bot from a server if you administer that server.
      </p>
    </section>

    <section>
      <h2>8. Children&apos;s Privacy</h2>
      <p>
        The giveaway bot is not directed to children under 13 or to anyone below the minimum age
        required to use Discord or GitHub in their country. We do not knowingly collect information
        from children who are not allowed to use these services.
      </p>
    </section>

    <section>
      <h2>9. International Processing</h2>
      <p>
        The giveaway service uses Discord, GitHub, and Railway. Your information may be processed
        in countries where those providers or EmuReady infrastructure operate.
      </p>
    </section>

    <section>
      <h2>10. Changes to This Policy</h2>
      <p>
        We may update this Privacy Policy when the giveaway service changes or when legal,
        platform, or operational requirements change. The latest version will be available at this
        page.
      </p>
    </section>

    <section>
      <h2>11. Contact</h2>
      <p>
        For questions, deletion requests, or privacy concerns about the giveaway bot, contact
        EmuReady staff in the EmuReady Discord server.
      </p>
    </section>
  </main>
</body>
</html>`
