import { useState, useCallback, useMemo } from "react";
import { useIsMobile } from "./use-mobile";

import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface DialogOrDrawerProps {
  title?: string;
  description?: string;
  children?: React.ReactNode | ((close: () => void) => React.ReactNode);
}

interface DialogComponentProps {
  children?: React.ReactNode | ((close: () => void) => React.ReactNode);
  title?: string;
  description?: string;
}

export const useDialog = (defaultProps?: DialogOrDrawerProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const isMobile = useIsMobile();

  const open = useCallback(() => setIsOpen(true), []);
  const close = useCallback(() => setIsOpen(false), []);

  const Component = useMemo(() => {
    const DialogComponent = ({
      children,
      title,
      description,
    }: DialogComponentProps) => {
      // Props passed to Component override default props
      const resolvedTitle = title ?? defaultProps?.title;
      const resolvedDescription = description ?? defaultProps?.description;
      const resolvedChildren = children ?? defaultProps?.children;
      const content =
        typeof resolvedChildren === "function"
          ? resolvedChildren(close)
          : resolvedChildren;

      if (isMobile) {
        return (
          <Drawer open={isOpen} onOpenChange={setIsOpen}>
            <DrawerContent>
              <DrawerHeader>
                {resolvedTitle && <DrawerTitle>{resolvedTitle}</DrawerTitle>}
                {resolvedDescription && (
                  <p className="text-sm text-muted-foreground">
                    {resolvedDescription}
                  </p>
                )}
              </DrawerHeader>
              <div className="px-4 pb-4">{content}</div>
            </DrawerContent>
          </Drawer>
        );
      }

      return (
        <Dialog open={isOpen} onOpenChange={setIsOpen}>
          <DialogContent className="sm:max-w-106.25">
            <DialogHeader>
              {resolvedTitle && <DialogTitle>{resolvedTitle}</DialogTitle>}
              {resolvedDescription && (
                <p className="text-sm text-muted-foreground">
                  {resolvedDescription}
                </p>
              )}
            </DialogHeader>
            {content}
          </DialogContent>
        </Dialog>
      );
    };
    return DialogComponent;
  }, [
    isOpen,
    isMobile,
    close,
    defaultProps?.title,
    defaultProps?.description,
    defaultProps?.children,
  ]);

  return {
    Component,
    open,
    close,
    isOpen,
  };
};
