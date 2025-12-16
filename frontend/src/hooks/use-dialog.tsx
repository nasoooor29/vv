import { useState } from "react";
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
  children: React.ReactNode;
}

export const useDialog = (props: DialogOrDrawerProps) => {
  const [isOpen, setIsOpen] = useState(false);

  const isMobile = useIsMobile();
  const dialogContent = () => {
    if (isMobile) {
      return (
        <Drawer open={isOpen} onOpenChange={setIsOpen}>
          <DrawerContent>
            <DrawerHeader>
              {props.title && <DrawerTitle>{props.title}</DrawerTitle>}
              {props.description && (
                <p className="text-sm text-muted-foreground">
                  {props.description}
                </p>
              )}
            </DrawerHeader>
            <div className="px-4 pb-4">{props.children}</div>
          </DrawerContent>
        </Drawer>
      );
    }

    return (
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="sm:max-w-106.25">
          <DialogHeader>
            {props.title && <DialogTitle>{props.title}</DialogTitle>}
            {props.description && (
              <p className="text-sm text-muted-foreground">
                {props.description}
              </p>
            )}
          </DialogHeader>
          {props.children}
        </DialogContent>
      </Dialog>
    );
  };
  return {
    component: dialogContent,
    open: () => setIsOpen(true),
    close: () => setIsOpen(false),
    isOpen,
  };
};
