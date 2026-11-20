/** The local calendar date as YYYY-MM-DD. */
export const toIsoDate = (date: Date): string => {
  const year = date.getFullYear().toString().padStart(4, "0");
  const month = (date.getMonth() + 1).toString().padStart(2, "0");
  const day = date.getDate().toString().padStart(2, "0");
  return `${year}-${month}-${day}`;
};

// Both dates count as calendar days, so the time of day and a daylight saving
// change between them do not move the result.
export const daysBetween = (today: Date, reviewed: Date): number => {
  const msPerDay = 24 * 60 * 60 * 1000;
  const a = Date.UTC(today.getFullYear(), today.getMonth(), today.getDate());
  const b = Date.UTC(
    reviewed.getFullYear(),
    reviewed.getMonth(),
    reviewed.getDate(),
  );
  return Math.round((a - b) / msPerDay);
};
