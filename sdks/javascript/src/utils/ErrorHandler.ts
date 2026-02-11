/**
 * Error Handler - Centralized error management
 */

export class TaskirXError extends Error {
  constructor(
    public code: string,
    public status: number,
    public details?: any
  ) {
    super(code);
    this.name = 'TaskirXError';
  }
}

export class ErrorHandler {
  static fromResponse(status: number, errorData: any): TaskirXError {
    const code = errorData.code || errorData.message || 'UNKNOWN_ERROR';
    const message = errorData.message || 'An error occurred';

    let errorCode: string;
    switch (status) {
      case 400:
        errorCode = 'BAD_REQUEST';
        break;
      case 401:
        errorCode = 'UNAUTHORIZED';
        break;
      case 403:
        errorCode = 'FORBIDDEN';
        break;
      case 404:
        errorCode = 'NOT_FOUND';
        break;
      case 429:
        errorCode = 'RATE_LIMIT_EXCEEDED';
        break;
      case 500:
        errorCode = 'SERVER_ERROR';
        break;
      case 503:
        errorCode = 'SERVICE_UNAVAILABLE';
        break;
      default:
        errorCode = code;
    }

    return new TaskirXError(errorCode, status, { message, ...errorData });
  }

  static handle(error: any): void {
    if (error instanceof TaskirXError) {
      console.error(`[${error.code}] ${error.details?.message || error.message}`);
    } else {
      console.error('Unexpected error:', error);
    }
  }
}
