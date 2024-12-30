import { Injectable } from '@nestjs/common';
import { Curio } from 'src/models/curio';
import { v4 } from 'uuid';

const storage: Curio<any>[] = [];

@Injectable()
export class ApiService {
  getCurio(id: string): Curio<any> | Error {
    return storage.find((value) => value.identifier.id === id);
  }

  searchCurios(body: any): any {
    return body;
  }

  createCurio(curio: Partial<Curio<any>>): Curio<any> | Error {
    const id = v4();
    const createdCurio = {
      identifier: {
        id: id,
        barcode: '',
        qrCode: '',
      },
      media: curio.media,
      addedDate: curio.addedDate,
      categorization: curio.categorization,
      borrowed: [],
      returned: [],
    };
    storage.push(createdCurio);
    return createdCurio;
  }

  updateCurio(body: any): Curio<any> | Error {
    return body;
  }

  deleteCurio(id: string): string | Error {
    const index = storage.findIndex((value) => value.identifier.id === id);
    if (index < 0) {
      return new Error();
    }
    storage.reduce((_prev, curr, i) => {
      if (i === index) {
        return;
      }
      return curr;
    });
    return id;
  }
}
