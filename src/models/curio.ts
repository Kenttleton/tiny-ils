import { Categorization } from './categorization';

export enum MediaType {
  THING,
  BOOK,
  VIDEO,
  AUDIO,
  GAME,
}

export enum FormatType {
  DIGITAL,
  PHYSICAL,
}

export type Format = {
  type: FormatType;
  description: string;
};

export type Thing = {
  label: string;
  description: string;
};

export type Book = {
  isbn: string;
  series: string;
  title: string;
  author: string;
  pages: number;
  summary: string;
  publishedBy: string;
  publishedDate: string;
};

export type Video = {
  director: string;
};

export type Audio = {
  album: string;
  artist: string;
  title: string;
};

export type Game = {
  series: string;
  title: string;
  developer: string;
  publisher: string;
  releaseDate: string;
};

export type Borrowed = {
  by: string;
  date: string;
};

export type Returned = {
  by: string;
  date: string;
};

export type Identifier = {
  id: string;
  barcode: string;
  qrCode: string;
};

export type CurioType = Thing | Book | Video | Audio | Game;

export type Media<M extends CurioType> = {
  type: MediaType;
  format: Format;
  metadata: M;
};

export type Curio<M extends CurioType> = {
  identifier: Identifier;
  media: Media<M>;
  addedDate: string;
  categorization: Categorization;
  borrowed: Borrowed[];
  returned: Returned[];
};
