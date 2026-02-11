// Type definitions to suppress errors when node_modules are missing
declare var process: {
  env: {
    [key: string]: string | undefined;
    NEXT_PUBLIC_BACKEND_URL?: string;
  }
};

declare var Buffer: {
  from: (str: string, encoding?: string) => any;
  [key: string]: any;
};

// Global React namespace
declare namespace React {
    type ElementType = any;
    type ReactNode = any;
    type FC<P = {}> = (props: P) => any;
    type Context<T> = { Provider: any, Consumer: any };
    type ChangeEvent<T> = any;
    type FormEvent<T = any> = any;
    type MouseEvent<T = any> = any;
    type KeyboardEvent<T = any> = any;
    type CSSProperties = any;
    
    type HTMLAttributes<T> = any;
    type ButtonHTMLAttributes<T> = any;
    type InputHTMLAttributes<T> = any;
    type TextareaHTMLAttributes<T> = any;
    type LabelHTMLAttributes<T> = any;
    type SelectHTMLAttributes<T> = any;
    type TableHTMLAttributes<T> = any;
    type TdHTMLAttributes<T> = any;
    type ThHTMLAttributes<T> = any;
    
    type ForwardedRef<T> = any;
    type ReactElement<P = any, T = any> = any;
    type RefObject<T> = { current: T | null };

    // Hooks
    function useState<T>(initialState: T | (() => T)): [T, (newState: T | ((prev: T) => T)) => void];
    function useEffect(effect: () => void | (() => void), deps?: any[]): void;
    function useRef<T>(initialValue?: T): { current: T };
    function createContext<T>(defaultValue: T): Context<T>;
    function useContext<T>(context: Context<T>): T;
    function useMemo<T>(factory: () => T, deps: any[] | undefined): T;
    function useCallback<T extends (...args: any[]) => any>(callback: T, deps: any[]): T;
    
    // API
    function forwardRef<T, P>(render: (props: P, ref: ForwardedRef<T>) => any): any;
    function isValidElement(object: any): boolean;
    function cloneElement(element: any, props?: any, ...children: any[]): any;
    function memo<T>(component: T): T;
}

declare module 'react' {
  export = React;
}

declare module 'next/link' {
  const Link: any;
  export default Link;
}

declare module 'next/navigation' {
  export const useRouter: any;
  export const usePathname: any;
  export const useSearchParams: any;
  export const redirect: any;
  export const notFound: any;
}

declare module 'next/constants' {
  export const PHASE_PRODUCTION_BUILD: string;
  export const PHASE_EXPORT: string;
  export const PHASE_DEVELOPMENT_SERVER: string;
  export const PHASE_TEST: string;
  const constants: any;
  export default constants;
}

declare module 'axios' {
  const axios: any;
  export default axios;
}

declare module 'lucide-react';

declare module 'next/server' {
  export class NextResponse {
    constructor(body?: any, init?: any);
    static json(body?: any, init?: any): any;
    static redirect(url: string | any, init?: any): any;
    static next(): any;
  }
  export type NextRequest = any;
}

declare module 'next' {
  export type Metadata = {
    title?: string;
    description?: string;
    keywords?: string[] | string;
    [key: string]: any;
  };
}

declare module '*.css';
declare module '*.module.css';
declare module '*.scss';
declare module '*.sass';

declare namespace JSX {
  interface IntrinsicElements {
    [elemName: string]: any;
  }
  interface IntrinsicAttributes {
    key?: any;
    ref?: any;
  }
}
