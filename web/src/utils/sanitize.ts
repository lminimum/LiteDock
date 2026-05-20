/**
 * Removes shell metacharacters (| ; $ ` \ " ') from input text.
 * Used to sanitize user input before passing to external systems.
 */
export function stripShellChars(text: string): string {
  return text.replace(/[|;$`\\"']/g, '')
}
