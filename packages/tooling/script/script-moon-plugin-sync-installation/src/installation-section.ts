export const installationStart = "<!-- xonovex:installation:start -->";
export const installationEnd = "<!-- xonovex:installation:end -->";

export const replaceInstallation = (text: string, block: string): string => {
  const marked = installationStart + "\n\n" + block + "\n\n" + installationEnd;
  const begins = text.indexOf(installationStart);
  const ends = text.indexOf(installationEnd);
  if (
    (begins === -1) !== (ends === -1) ||
    (begins !== -1 &&
      (ends < begins ||
        text.includes(installationStart, begins + installationStart.length) ||
        text.includes(installationEnd, ends + installationEnd.length)))
  ) {
    throw new Error("Installation markers must form one complete block.");
  }
  if (begins !== -1)
    return (
      text.slice(0, begins) + marked + text.slice(ends + installationEnd.length)
    );
  const heading = /^## (?:Install|Installation)\s*$/mu.exec(text);
  if (heading?.index !== undefined) {
    const remainder = text.slice(heading.index + heading[0].length);
    const nextHeading = /^## /mu.exec(remainder);
    const next =
      nextHeading?.index === undefined
        ? text.length
        : heading.index + heading[0].length + nextHeading.index;
    return text.slice(0, heading.index) + marked + "\n\n" + text.slice(next);
  }
  return text.trimEnd() + "\n\n" + marked + "\n";
};
