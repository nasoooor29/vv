import { base, detailed } from "./general";
import { z } from "zod";

const isoFileInfoSchema = z.object({
  name: z.string(),
  size: z.number(),
  modified: z.string().or(z.instanceof(Date)),
  is_directory: z.boolean(),
});

const isoUploadResponseSchema = z.object({
  name: z.string(),
  size: z.number(),
  modified: z.string().or(z.instanceof(Date)),
  message: z.string(),
});

const isoDeleteResponseSchema = z.object({
  message: z.string(),
});

export const isoRouter = {
  // List all ISO files
  listISOs: base
    .route({
      method: "GET",
      path: "/iso",
    })
    .output(isoFileInfoSchema.array()),

  // Get ISO file info
  getISOInfo: base
    .route({
      method: "GET",
      path: "/iso/{filename}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { filename: z.string() },
      }),
    )
    .output(isoFileInfoSchema),

  // Upload ISO file
  uploadISO: base
    .route({
      method: "POST",
      path: "/iso",
    })
    .input(
      z.object({
        file: z.instanceof(File),
      }),
    )
    .output(isoUploadResponseSchema),

  // Delete ISO file
  deleteISO: base
    .route({
      method: "DELETE",
      path: "/iso/{filename}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { filename: z.string() },
      }),
    )
    .output(isoDeleteResponseSchema),
};
