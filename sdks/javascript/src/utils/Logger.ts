/**
 * Logger - Centralized logging utility
 */

export class Logger {
  private isDebugEnabled: boolean;

  constructor(debugEnabled: boolean = false) {
    this.isDebugEnabled = debugEnabled;
  }

  setDebug(enabled: boolean): void {
    this.isDebugEnabled = enabled;
  }

  debug(message: string, ...args: any[]): void {
    if (this.isDebugEnabled) {
      console.log(`[TaskirX DEBUG] ${message}`, ...args);
    }
  }

  info(message: string, ...args: any[]): void {
    console.info(`[TaskirX INFO] ${message}`, ...args);
  }

  warn(message: string, ...args: any[]): void {
    console.warn(`[TaskirX WARN] ${message}`, ...args);
  }

  error(message: string, ...args: any[]): void {
    console.error(`[TaskirX ERROR] ${message}`, ...args);
  }
}
