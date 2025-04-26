import { ComponentFixture, TestBed } from '@angular/core/testing';

import { LinkLcPageComponent } from './link-lc-page.component';

describe('LinkLcPageComponent', () => {
  let component: LinkLcPageComponent;
  let fixture: ComponentFixture<LinkLcPageComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LinkLcPageComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(LinkLcPageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
