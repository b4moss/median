export class MIMEError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "MIMEError";
  }
}

export const ErrMIMEDenied = new MIMEError("mime: denied");
export const ErrMIMEEmpty = new MIMEError("mime: empty");

/** Validate declared MIME against allow/deny lists. Deny always wins; empty allow accepts all non-denied. */
export function checkMIME(declared: string, allow: string[] = [], deny: string[] = []): void {
  const mime = declared.trim();
  if (!mime) {
    throw new MIMEError("mime: empty");
  }
  for (const d of deny) {
    if (d.trim().toLowerCase() === mime.toLowerCase()) {
      throw new MIMEError("mime: denied");
    }
  }
  if (allow.length === 0) {
    return;
  }
  for (const a of allow) {
    if (a.trim().toLowerCase() === mime.toLowerCase()) {
      return;
    }
  }
  throw new MIMEError("mime: denied");
}
