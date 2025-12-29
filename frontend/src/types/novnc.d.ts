declare module 'novnc/core/rfb' {
  export class RFB {
    constructor(target: HTMLCanvasElement | OffscreenCanvas, url: string, options?: RFBOptions);
    addEventListener(event: string, handler: (e: any) => void): void;
    removeEventListener(event: string, handler: (e: any) => void): void;
    disconnect(): void;
    focus(): void;
    blur(): void;
    clipViewport: boolean;
    dragViewport: boolean;
    resizeSession: boolean;
    showDotCursor: boolean;
    viewOnly: boolean;
    scaleViewport: boolean;
    desktopName: string;
    capabilities: Capabilities;
  }

  export interface RFBOptions {
    credentials?: {
      username: string;
      password: string;
      target: string;
    };
  }

  export interface Capabilities {
    power: boolean;
  }
}
