import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ProblemFilterComponent } from './problem-filter.component';

describe('ProblemFilterComponent', () => {
  let component: ProblemFilterComponent;
  let fixture: ComponentFixture<ProblemFilterComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ProblemFilterComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ProblemFilterComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
