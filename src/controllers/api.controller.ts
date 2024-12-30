import {
  Body,
  Controller,
  Delete,
  Get,
  Param,
  Post,
  Put,
  Query,
} from '@nestjs/common';
import { ApiService } from '../services/api.service';
import { FormatType, MediaType } from 'src/models/curio';
import { enumToArray } from 'src/utilities/enum';

@Controller('api')
export class ApiController {
  constructor(private readonly apiService: ApiService) {}

  @Get('curios/:id')
  getCurio(@Param() params: any): any {
    return this.apiService.getCurio(params.id);
  }

  @Get('curios')
  searchCurios(@Query() query: any): any {
    return this.apiService.searchCurios(query);
  }

  @Post('curios')
  postCurios(@Body() body: any): any {
    return this.apiService.createCurio(body);
  }

  @Put('curios')
  putCurios(@Body() body: any): any {
    return this.apiService.updateCurio(body);
  }

  @Delete('curios')
  deleteCurios(@Body() body: any): any {
    return this.apiService.deleteCurio(body);
  }

  @Get('types')
  getAllTypes(): any {
    return {
      media: this.getMediaTypes(),
      formats: this.getFormatTypes(),
    };
  }

  @Get('types/media')
  getMediaTypes(): { key: string; value: typeof MediaType }[] {
    return enumToArray(MediaType);
  }

  @Get('types/formats')
  getFormatTypes(): { key: string; value: typeof FormatType }[] {
    return enumToArray(FormatType);
  }
}
