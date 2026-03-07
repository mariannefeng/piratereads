import posthog from "posthog-js";

posthog.init(process.env.NEXT_PUBLIC_POSTHOG_TOKEN!, {
  api_host: process.env.NEXT_PUBLIC_POSTHOG_HOST,
  person_profiles: "identified_only",
});
