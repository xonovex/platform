/**
 * The variables an evaluator reads its defaults from and hands to the harness it
 * starts. The entry file passes the process environment; a unit spec names its own.
 */
export type Environment = Readonly<Record<string, string | undefined>>;
