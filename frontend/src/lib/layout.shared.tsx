import type { BaseLayoutProps } from "fumadocs-ui/layouts/shared";
import Logo from "@/app/icon.svg";
import Image from "next/image";

// fill this with your actual GitHub info, for example:
export const gitConfig = {
  user: "mariannefeng",
  repo: "piratereads",
  branch: "main",
};

export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: (
        <>
          <Image
            src={Logo}
            alt="piratereads"
            width={16}
            height={16}
            className="dark:invert"
          />
          piratereads
        </>
      ),
    },
    githubUrl: `https://github.com/${gitConfig.user}/${gitConfig.repo}`,
  };
}
